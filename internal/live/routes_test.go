package live_test

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/krocon/go-gin-demo-app/internal/server"
	"github.com/krocon/go-gin-demo-app/internal/tasks"
)

func init() { gin.SetMode(gin.TestMode) }

// startServer runs the app on a random local port. WebSockets need a real
// network connection, so the in-memory httptest.ResponseRecorder is not enough here.
func startServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(server.NewRouter(tasks.NewInMemoryRepository(), tasks.NewEventBus()))
	t.Cleanup(srv.Close)
	return srv
}

func dial(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/tasks"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	return conn
}

func readEvent(t *testing.T, conn *websocket.Conn) tasks.Event {
	t.Helper()
	var e tasks.Event
	if err := conn.ReadJSON(&e); err != nil {
		t.Fatal(err)
	}
	return e
}

func TestTextFrameCreatesATaskAndIsBroadcastBack(t *testing.T) {
	conn := dial(t, startServer(t))

	_ = conn.WriteMessage(websocket.TextMessage, []byte("Created via WebSocket"))
	e := readEvent(t, conn)

	if e.Type != tasks.EventCreated || *e.Task != (tasks.Task{ID: 1, Title: "Created via WebSocket"}) {
		t.Fatalf("got %+v", e)
	}
}

func TestRESTDeleteIsPushedToWebSocketClients(t *testing.T) {
	srv := startServer(t)
	conn := dial(t, srv)

	// Creating a task first guarantees the socket is subscribed before we delete.
	_ = conn.WriteMessage(websocket.TextMessage, []byte("Delete me"))
	created := readEvent(t, conn)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/tasks/1", nil)
	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	}

	if e := readEvent(t, conn); e.Type != tasks.EventDeleted || e.ID != created.Task.ID {
		t.Fatalf("got %+v", e)
	}
}

func TestRESTCreateIsPushedAsServerSentEvent(t *testing.T) {
	srv := startServer(t)

	// The handler flushes headers right after subscribing, so once Get returns we are subscribed.
	client := &http.Client{Timeout: 5 * time.Second} // don't hang forever if no event arrives
	resp, err := client.Get(srv.URL + "/tasks/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("Content-Type = %q", ct)
	}

	_, err = http.Post(srv.URL+"/tasks", "application/json", strings.NewReader(`{"title":"Via SSE"}`))
	if err != nil {
		t.Fatal(err)
	}

	// Expect: "event:created" then "data:{...}"
	lines := bufio.NewScanner(resp.Body)
	var eventName string
	for lines.Scan() {
		line := lines.Text()
		if name, ok := strings.CutPrefix(line, "event:"); ok {
			eventName = strings.TrimSpace(name)
		}
		if data, ok := strings.CutPrefix(line, "data:"); ok {
			var e tasks.Event
			if err := json.Unmarshal([]byte(data), &e); err != nil {
				t.Fatal(err)
			}
			if eventName != tasks.EventCreated || e.Task.Title != "Via SSE" {
				t.Fatalf("got event %q %+v", eventName, e)
			}
			return
		}
	}
	t.Fatal("stream ended without an event")
}
