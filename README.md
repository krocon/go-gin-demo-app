# Go + Gin Demo App

A small, self-contained REST API built with [Gin](https://gin-gonic.com) and Go.
It is the Go counterpart of [kotlin-ktor-demo-app](https://github.com/krocon/kotlin-ktor-demo-app)
and the reference example for a series of one-minute videos — every file
focuses on **one** concept, so each video can point to exactly one place in the code.

## Quick start

Requirements: Go 1.24 or newer.

```bash
go mod tidy                 # first time only: download dependencies, complete go.sum
go run ./cmd/server         # start the server on http://localhost:8080
go test ./...               # run all tests
go build -o bin/server ./cmd/server   # build a single static binary
```

Try it:

```bash
curl http://localhost:8080/tasks
curl -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"title":"Try Gin"}'
```

Or open [`requests.http`](requests.http) in GoLand / IntelliJ IDEA and click the ▶ icons.

## Endpoints

| Method | Path            | Description                              |
|--------|-----------------|------------------------------------------|
| GET    | `/`             | Hello world                              |
| GET    | `/health`       | Health check                             |
| GET    | `/tasks`        | List tasks (optional `?done=true/false`) |
| GET    | `/tasks/{id}`   | Get one task                             |
| POST   | `/tasks`        | Create a task                            |
| PATCH  | `/tasks/{id}`   | Update title and/or done                 |
| DELETE | `/tasks/{id}`   | Delete a task                            |
| WS     | `/ws/tasks`     | Live task events; send text to create    |
| GET    | `/tasks/events` | Live task events as Server-Sent Events   |
| GET    | `/live/`        | Browser demo page for the live features  |

Errors are always JSON: `{"message": "Task 42 not found"}` with the matching HTTP status
(400 for validation / bad input, 404 for unknown tasks, 500 for everything unexpected).

## Live updates (WebSockets & SSE)

Open <http://localhost:8080/live/> in two browser windows. Add, check or delete a task in one
window – the other one updates instantly.

Every change made through the REST API or the WebSocket is published to an `EventBus`
(plain Go channels) and pushed to all connected clients as JSON:

```json
{"type":"created","task":{"id":4,"title":"Try Gin","done":false}}
```

From the terminal:

```bash
curl -N http://localhost:8080/tasks/events   # SSE stream, keep open
websocat ws://localhost:8080/ws/tasks          # WebSocket, type a title + Enter
```

## Project map (one topic per file)

| Topic                              | File                                                                         |
|------------------------------------|------------------------------------------------------------------------------|
| Starting the server, graceful shutdown | [`cmd/server/main.go`](cmd/server/main.go)                               |
| Routing basics / wiring            | [`internal/server/router.go`](internal/server/router.go)                     |
| REST routes / CRUD                 | [`internal/tasks/routes.go`](internal/tasks/routes.go)                       |
| Structs, JSON tags, DTOs           | [`internal/tasks/task.go`](internal/tasks/task.go)                           |
| Error handling middleware          | [`internal/server/errors.go`](internal/server/errors.go)                     |
| Errors with HTTP status            | [`internal/httperr/httperr.go`](internal/httperr/httperr.go)                 |
| Request validation (`binding` tags)| [`internal/server/validation.go`](internal/server/validation.go)             |
| Live endpoints (WS + SSE)          | [`internal/live/routes.go`](internal/live/routes.go)                         |
| Pub/sub with channels              | [`internal/tasks/event_bus.go`](internal/tasks/event_bus.go)                 |
| Tagged JSON events                 | [`internal/tasks/event.go`](internal/tasks/event.go)                         |
| Embedding static files (`go:embed`)| [`web/embed.go`](web/embed.go)                                               |
| Browser WebSocket client           | [`web/static/index.html`](web/static/index.html)                             |
| Testing WebSockets & SSE           | [`internal/live/routes_test.go`](internal/live/routes_test.go)               |
| Logging middleware (`log/slog`)    | [`internal/server/logging.go`](internal/server/logging.go)                   |
| Repository pattern (interface)     | [`internal/tasks/repository.go`](internal/tasks/repository.go)               |
| Thread-safe in-memory storage      | [`internal/tasks/memory_repository.go`](internal/tasks/memory_repository.go) |
| Testing routes (`httptest`)        | [`internal/server/router_test.go`](internal/server/router_test.go)           |
| Unit testing                       | [`internal/tasks/memory_repository_test.go`](internal/tasks/memory_repository_test.go) |
| Dependencies                       | [`go.mod`](go.mod)                                                           |
| Docker                             | [`Dockerfile`](Dockerfile)                                                   |

## Ktor → Gin cheat sheet

| Ktor (Kotlin)                         | Gin (Go)                                               |
|---------------------------------------|--------------------------------------------------------|
| `embeddedServer(Netty)`               | `http.Server{Handler: gin.New()}`                      |
| Plugin (`install(...)`)               | Middleware (`r.Use(...)`)                              |
| `StatusPages`                         | `c.Error(err)` + `ErrorHandler` middleware             |
| `RequestValidation`                   | `binding:"..."` struct tags (go-playground/validator)  |
| `kotlinx.serialization`               | `encoding/json` + `json:"..."` tags                    |
| Nullable fields (`String?`)           | Pointers (`*string`)                                   |
| `suspend fun` + coroutines            | `context.Context` + goroutines                         |
| `SharedFlow`                          | Channels + `sync.RWMutex` (`EventBus`)                 |
| Sealed class events                   | One struct with a `type` field + `omitempty`           |
| `ConcurrentHashMap` / `AtomicLong`    | `map` guarded by `sync.RWMutex`                        |
| `staticResources(...)`                | `//go:embed` + `r.StaticFS(...)`                       |
| `testApplication { }`                 | `httptest.NewRecorder()` / `httptest.NewServer()`      |
| `CallLogging`                         | Custom middleware with `log/slog`                      |

## Docker

```bash
docker build -t gin-demo .
docker run -p 8080:8080 gin-demo
```

The image uses a multi-stage build and ships only the static binary on a distroless base (~15 MB).

## Configuration

| Variable | Default | Description      |
|----------|---------|------------------|
| `PORT`   | `8080`  | HTTP server port |
