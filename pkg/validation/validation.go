// Package validation chuyển lỗi binding/validator của Gin thành *apperror.AppError
// với thông báo dễ đọc, ví dụ: {"errors": {"body": ["email can't be blank"]}}.
package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"conduit/pkg/apperror"
)

// RegisterJSONTagNames để validator báo lỗi theo tên field JSON ("email")
// thay vì tên field Go ("Email"). Gọi một lần lúc khởi động.
func RegisterJSONTagNames() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		if name == "" {
			return f.Name
		}
		return name
	})
}

// Translate chuyển lỗi từ c.ShouldBindJSON thành lỗi 422.
func Translate(err error) *apperror.AppError {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		msgs := make([]string, 0, len(verrs))
		for _, fe := range verrs {
			msgs = append(msgs, message(fe))
		}
		return apperror.Unprocessable(msgs...)
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		field := typeErr.Field
		if i := strings.LastIndex(field, "."); i >= 0 {
			field = field[i+1:]
		}
		return apperror.Unprocessable(fmt.Sprintf("%s must be a %s", field, typeErr.Type.String()))
	}

	if errors.Is(err, io.EOF) {
		return apperror.Unprocessable("request body can't be empty")
	}
	return apperror.Unprocessable("request body is not valid JSON")
}

func message(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return field + " can't be blank"
	case "email":
		return field + " is invalid"
	case "min":
		return fmt.Sprintf("%s is too short (minimum is %s characters)", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s is too long (maximum is %s characters)", field, fe.Param())
	default:
		return field + " is invalid"
	}
}
