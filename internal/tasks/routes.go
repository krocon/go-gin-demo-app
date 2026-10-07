package tasks

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/krocon/go-gin-demo-app/internal/httperr"
)

// RegisterRoutes adds the REST endpoints for tasks. Every change is published to the EventBus:
//
//	GET    /tasks            list all tasks (optional ?done=true|false)
//	GET    /tasks/{id}       get one task
//	POST   /tasks            create a task
//	PATCH  /tasks/{id}       update title and/or done
//	DELETE /tasks/{id}       delete a task
func RegisterRoutes(r gin.IRouter, repo Repository, bus *EventBus) {
	h := &handler{repo: repo, bus: bus}

	g := r.Group("/tasks")
	g.GET("", h.list)
	g.GET("/:id", h.get)
	g.POST("", h.create)
	g.PATCH("/:id", h.update)
	g.DELETE("/:id", h.delete)
}

type handler struct {
	repo Repository
	bus  *EventBus
}

func (h *handler) list(c *gin.Context) {
	done, err := doneFilter(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	tasks, err := h.repo.FindAll(c.Request.Context(), done)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *handler) get(c *gin.Context) {
	id, err := taskID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	task, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(notFound(err, id))
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *handler) create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil { // decodes JSON *and* runs `binding` validation
		_ = c.Error(err)
		return
	}
	task, err := h.repo.Create(c.Request.Context(), req.Title)
	if err != nil {
		_ = c.Error(err)
		return
	}
	h.bus.Publish(Created(task))
	c.Header("Location", fmt.Sprintf("/tasks/%d", task.ID))
	c.JSON(http.StatusCreated, task)
}

func (h *handler) update(c *gin.Context) {
	id, err := taskID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}
	task, err := h.repo.Update(c.Request.Context(), id, req)
	if err != nil {
		_ = c.Error(notFound(err, id))
		return
	}
	h.bus.Publish(Updated(task))
	c.JSON(http.StatusOK, task)
}

func (h *handler) delete(c *gin.Context) {
	id, err := taskID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		_ = c.Error(notFound(err, id))
		return
	}
	h.bus.Publish(Deleted(id))
	c.Status(http.StatusNoContent)
}

func taskID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return 0, httperr.BadRequest("id must be a number")
	}
	return id, nil
}

// doneFilter reads ?done=true|false. Missing means "no filter" (nil).
func doneFilter(c *gin.Context) (*bool, error) {
	raw, ok := c.GetQuery("done")
	if !ok {
		return nil, nil
	}
	switch raw {
	case "true":
		v := true
		return &v, nil
	case "false":
		v := false
		return &v, nil
	default:
		return nil, httperr.BadRequest("done must be true or false")
	}
}

// notFound translates the repository's ErrNotFound into a 404 with a helpful message.
func notFound(err error, id int64) error {
	if errors.Is(err, ErrNotFound) {
		return httperr.NotFound(fmt.Sprintf("Task %d not found", id))
	}
	return err
}
