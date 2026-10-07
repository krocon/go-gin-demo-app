// Package tasks contains the task domain: model, storage, events and REST routes.
package tasks

// Task is a single to-do item. Struct tags control the JSON field names.
type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// CreateTaskRequest is the body of POST /tasks.
// The `binding` tag is checked by gin when the body is bound (see server/validation.go).
type CreateTaskRequest struct {
	Title string `json:"title" binding:"notblank,max=100"`
}

// UpdateTaskRequest is the body of PATCH /tasks/{id}.
// Pointers distinguish "not sent" (nil) from a zero value — only set fields get updated.
type UpdateTaskRequest struct {
	Title *string `json:"title" binding:"omitempty,notblank,max=100"`
	Done  *bool   `json:"done"`
}
