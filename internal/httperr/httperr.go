// Package httperr defines errors that carry an HTTP status code.
// Handlers return them via c.Error(...); server.ErrorHandler turns them into JSON.
package httperr

import "net/http"

// Error is an error with an HTTP status and a client-facing message.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

// BadRequest creates a 400 error.
func BadRequest(msg string) *Error { return &Error{Status: http.StatusBadRequest, Message: msg} }

// NotFound creates a 404 error.
func NotFound(msg string) *Error { return &Error{Status: http.StatusNotFound, Message: msg} }

// Response is the JSON body of every error response: {"message": "..."}.
type Response struct {
	Message string `json:"message"`
}
