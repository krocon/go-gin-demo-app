// Package live provides the real-time endpoints:
//
//	WS  /ws/tasks      two-way: receive all task events, send a text frame to create a task
//	SSE /tasks/events  one-way: receive all task events as Server-Sent Events
package live

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/krocon/go-gin-demo-app/internal/tasks"
)

// upgrader turns an HTTP request into a WebSocket connection.
// By default it only accepts same-origin browser requests (and non-browser clients).
var upgrader = websocket.Upgrader{}

func RegisterRoutes(r gin.IRouter, repo tasks.Repository, bus *tasks.EventBus) {
	r.GET("/ws/tasks", webSocketHandler(repo, bus))
	r.GET("/tasks/events", sseHandler(bus))
}

func webSocketHandler(repo tasks.Repository, bus *tasks.EventBus) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return // Upgrade already wrote an HTTP error response
		}
		defer conn.Close()

		events, unsubscribe := bus.Subscribe()
		defer unsubscribe()

		// Server -> client: forward every task event as a JSON text frame.
		// A gorilla connection allows only one concurrent writer, so all writes happen here.
		writerDone := make(chan struct{})
		go func() {
			defer close(writerDone)
			for e := range events { // ends when unsubscribe() closes the channel
				if err := conn.WriteJSON(e); err != nil {
					return
				}
			}
		}()

		// Client -> server: every text frame becomes a new task.
		for {
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				break // client closed the connection
			}
			if msgType != websocket.TextMessage {
				continue
			}
			title := strings.TrimSpace(string(data))
			if title == "" {
				continue
			}
			task, err := repo.Create(c.Request.Context(), title)
			if err != nil {
				slog.Error("create via websocket", "error", err)
				continue
			}
			bus.Publish(tasks.Created(task))
		}

		unsubscribe()
		<-writerDone
	}
}

func sseHandler(bus *tasks.EventBus) gin.HandlerFunc {
	return func(c *gin.Context) {
		events, unsubscribe := bus.Subscribe()
		defer unsubscribe()

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Status(http.StatusOK)
		c.Writer.Flush() // send headers now, so the client knows the stream is open

		for {
			select {
			case e := <-events:
				c.SSEvent(e.Type, e) // writes "event: <type>\ndata: <json>\n\n"
				c.Writer.Flush()
			case <-c.Request.Context().Done():
				return // client disconnected (or server shutting down)
			}
		}
	}
}
