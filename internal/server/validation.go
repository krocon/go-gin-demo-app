package server

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var registerOnce sync.Once

// registerValidators teaches gin's validator (go-playground/validator) two things:
//   - a custom "notblank" rule used in the `binding:"..."` struct tags
//   - to report JSON field names ("title") instead of Go field names ("Title")
func registerValidators() {
	registerOnce.Do(func() {
		v, ok := binding.Validator.Engine().(*validator.Validate)
		if !ok {
			return
		}
		_ = v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
			return strings.TrimSpace(fl.Field().String()) != ""
		})
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "" || name == "-" {
				return f.Name
			}
			return name
		})
	})
}

// validationMessage turns validator errors into readable text,
// e.g. "title must not be blank".
func validationMessage(errs validator.ValidationErrors) string {
	msgs := make([]string, 0, len(errs))
	for _, fe := range errs {
		switch fe.Tag() {
		case "notblank", "required":
			msgs = append(msgs, fmt.Sprintf("%s must not be blank", fe.Field()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid (%s)", fe.Field(), fe.Tag()))
		}
	}
	return strings.Join(msgs, ", ")
}
