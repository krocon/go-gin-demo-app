package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/krocon/go-gin-demo-app/internal/httperr"
	"github.com/krocon/go-gin-demo-app/internal/tasks"
)

func init() { gin.SetMode(gin.TestMode) }

// do sends one request through the router in memory – no real network involved.
func do(t *testing.T, router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("invalid JSON %q: %v", rec.Body.String(), err)
	}
	return v
}

// setup starts the app with the given (or an empty) repository.
func setup(repo ...*tasks.InMemoryRepository) http.Handler {
	r := tasks.NewInMemoryRepository()
	if len(repo) > 0 {
		r = repo[0]
	}
	return NewRouter(r, tasks.NewEventBus())
}

func TestRootReturnsHello(t *testing.T) {
	rec := do(t, setup(), http.MethodGet, "/", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Hello, Gin!") || !strings.Contains(rec.Body.String(), `href="/live/"`) {
		t.Fatalf("unexpected body %q", rec.Body.String())
	}
}

func TestHealthCheckIsUp(t *testing.T) {
	rec := do(t, setup(), http.MethodGet, "/health", "")

	if got := decode[map[string]string](t, rec); rec.Code != http.StatusOK || got["status"] != "UP" {
		t.Fatalf("got %d %v", rec.Code, got)
	}
}

func TestCreateAndFetchATask(t *testing.T) {
	router := setup()

	created := do(t, router, http.MethodPost, "/tasks", `{"title":"Write tests"}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", created.Code, created.Body)
	}
	if loc := created.Header().Get("Location"); loc != "/tasks/1" {
		t.Fatalf("Location = %q", loc)
	}

	got := decode[tasks.Task](t, do(t, router, http.MethodGet, "/tasks/1", ""))
	if want := (tasks.Task{ID: 1, Title: "Write tests"}); got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestListTasksFilteredByDone(t *testing.T) {
	repo := tasks.NewInMemoryRepository()
	ctx := context.Background()
	_, _ = repo.Create(ctx, "Open task")
	closed, _ := repo.Create(ctx, "Closed task")
	done := true
	_, _ = repo.Update(ctx, closed.ID, tasks.UpdateTaskRequest{Done: &done})

	got := decode[[]tasks.Task](t, do(t, setup(repo), http.MethodGet, "/tasks?done=true", ""))

	if len(got) != 1 || got[0].Title != "Closed task" {
		t.Fatalf("got %+v", got)
	}
}

func TestInvalidDoneFilterReturns400(t *testing.T) {
	rec := do(t, setup(), http.MethodGet, "/tasks?done=yes", "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestPatchMarksATaskAsDone(t *testing.T) {
	repo := tasks.NewInMemoryRepository()
	task, _ := repo.Create(context.Background(), "Ship it")

	rec := do(t, setup(repo), http.MethodPatch, "/tasks/1", `{"done":true}`)

	if rec.Code != http.StatusOK || !decode[tasks.Task](t, rec).Done {
		t.Fatalf("got %d %s for task %d", rec.Code, rec.Body, task.ID)
	}
}

func TestPatchWithBlankTitleIsRejected(t *testing.T) {
	repo := tasks.NewInMemoryRepository()
	_, _ = repo.Create(context.Background(), "Keep me")

	rec := do(t, setup(repo), http.MethodPatch, "/tasks/1", `{"title":"  "}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
}

func TestDeleteRemovesATask(t *testing.T) {
	repo := tasks.NewInMemoryRepository()
	_, _ = repo.Create(context.Background(), "Temporary")
	router := setup(repo)

	if rec := do(t, router, http.MethodDelete, "/tasks/1", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete status %d", rec.Code)
	}
	if rec := do(t, router, http.MethodGet, "/tasks/1", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("get status %d", rec.Code)
	}
}

func TestBlankTitleIsRejected(t *testing.T) {
	rec := do(t, setup(), http.MethodPost, "/tasks", `{"title":"   "}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
	if msg := decode[httperr.Response](t, rec).Message; msg != "title must not be blank" {
		t.Fatalf("message %q", msg)
	}
}

func TestTooLongTitleIsRejected(t *testing.T) {
	body := `{"title":"` + strings.Repeat("x", 101) + `"}`

	rec := do(t, setup(), http.MethodPost, "/tasks", body)

	if msg := decode[httperr.Response](t, rec).Message; rec.Code != http.StatusBadRequest || msg != "title must be at most 100 characters" {
		t.Fatalf("got %d %q", rec.Code, msg)
	}
}

func TestInvalidJSONReturns400(t *testing.T) {
	rec := do(t, setup(), http.MethodPost, "/tasks", `{not json`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestUnknownTaskReturns404(t *testing.T) {
	rec := do(t, setup(), http.MethodGet, "/tasks/42", "")

	if msg := decode[httperr.Response](t, rec).Message; rec.Code != http.StatusNotFound || msg != "Task 42 not found" {
		t.Fatalf("got %d %q", rec.Code, msg)
	}
}

func TestNonNumericIDReturns400(t *testing.T) {
	if rec := do(t, setup(), http.MethodGet, "/tasks/abc", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestLivePageIsServed(t *testing.T) {
	rec := do(t, setup(), http.MethodGet, "/live/", "")

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Gin Live Tasks") {
		t.Fatalf("got %d", rec.Code)
	}
}
