// Package apperror định nghĩa lỗi nghiệp vụ mang theo HTTP status.
// Service trả về *AppError; middleware ErrorHandler chuyển nó thành JSON
// đúng format RealWorld: {"errors": {"body": ["..."]}}.
package apperror

import (
	"fmt"
	"net/http"
)

// BodyKey là key mặc định trong object "errors" của RealWorld spec.
const BodyKey = "body"

type AppError struct {
	Status int
	Errors map[string][]string
	// Err là nguyên nhân gốc (lỗi DB, lỗi thư viện...). Chỉ dùng để log,
	// KHÔNG BAO GIỜ trả ra client để tránh lộ thông tin nội bộ.
	Err error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("status %d: %v: %v", e.Status, e.Errors, e.Err)
	}
	return fmt.Sprintf("status %d: %v", e.Status, e.Errors)
}

func (e *AppError) Unwrap() error { return e.Err }

func New(status int, messages ...string) *AppError {
	if len(messages) == 0 {
		messages = []string{http.StatusText(status)}
	}
	return &AppError{Status: status, Errors: map[string][]string{BodyKey: messages}}
}

// Unprocessable — 422: dữ liệu đầu vào không hợp lệ.
func Unprocessable(messages ...string) *AppError {
	return New(http.StatusUnprocessableEntity, messages...)
}

// Unauthorized — 401: thiếu hoặc sai thông tin xác thực.
func Unauthorized(messages ...string) *AppError {
	return New(http.StatusUnauthorized, messages...)
}

// Forbidden — 403: đã xác thực nhưng không có quyền (vd: sửa bài của người khác).
func Forbidden(messages ...string) *AppError {
	return New(http.StatusForbidden, messages...)
}

// NotFound — 404: tài nguyên không tồn tại.
func NotFound(messages ...string) *AppError {
	return New(http.StatusNotFound, messages...)
}

// Internal — 500: lỗi hệ thống. Client chỉ thấy thông báo chung chung.
func Internal(err error) *AppError {
	e := New(http.StatusInternalServerError, "internal server error")
	e.Err = err
	return e
}
