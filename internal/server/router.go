// Package server wires middleware and routes together into one gin.Engine.
package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/krocon/go-gin-demo-app/internal/live"
	"github.com/krocon/go-gin-demo-app/internal/tasks"
	"github.com/krocon/go-gin-demo-app/web"
)

const rootPage = `<h1>Hello, Gin!</h1>
<p><a href="/live/">Open the live demo &rarr;</a></p>`

// NewRouter builds the complete application.
// Dependencies are parameters so tests can pass in fresh instances.
func NewRouter(repo tasks.Repository, bus *tasks.EventBus) *gin.Engine {
	registerValidators()

	r := gin.New() // gin.Default() would add gin's own logger + recovery; we bring our own.
	r.Use(RequestLogger(), Recovery(), ErrorHandler())

	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(rootPage))
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	tasks.RegisterRoutes(r, repo, bus)
	live.RegisterRoutes(r, repo, bus)

	// Browser demo page for the live endpoints: http://localhost:8080/live/
	r.StaticFS("/live", web.StaticFS())

	return r
}
