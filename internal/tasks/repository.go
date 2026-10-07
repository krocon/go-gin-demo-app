package tasks

import (
	"context"
	"errors"
)

// ErrNotFound is returned when no task with the given id exists.
var ErrNotFound = errors.New("task not found")

// Repository is the storage abstraction for tasks.
// Swap the implementation (e.g. for a database) without touching the routes.
type Repository interface {
	// FindAll lists tasks sorted by id; done == nil means "no filter".
	FindAll(ctx context.Context, done *bool) ([]Task, error)
	FindByID(ctx context.Context, id int64) (Task, error)
	Create(ctx context.Context, title string) (Task, error)
	Update(ctx context.Context, id int64, req UpdateTaskRequest) (Task, error)
	Delete(ctx context.Context, id int64) error
}
