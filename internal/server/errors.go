package server

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/krocon/go-gin-demo-app/internal/httperr"
)

// ErrorHandler maps errors collected with c.Error(...) to clean JSON error
// responses with the right HTTP status. Handlers just call c.Error(err) and return.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}
		status, msg := toStatus(c.Errors.Last().Err)
		c.JSON(status, httperr.Response{Message: msg})
	}
}

func toStatus(err error) (int, string) {
	var (
		httpErr    *httperr.Error
		validation validator.ValidationErrors
		syntaxErr  *json.SyntaxError
		typeErr    *json.UnmarshalTypeError
	)
	switch {
	case errors.As(err, &httpErr):
		return httpErr.Status, httpErr.Message
	case errors.As(err, &validation):
		return http.StatusBadRequest, validationMessage(validation)
	case errors.As(err, &syntaxErr), errors.As(err, &typeErr),
		errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return http.StatusBadRequest, "invalid JSON body"
	default:
		slog.Error("unhandled error", "error", err)
		return http.StatusInternalServerError, "Internal server error"
	}
}

// Recovery turns a panic in a handler into a 500 JSON response instead of a dropped connection.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("panic", "recovered", recovered)
		c.AbortWithStatusJSON(http.StatusInternalServerError, httperr.Response{Message: "Internal server error"})
	})
}
