package middlewares

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"conduit/dto"
	"conduit/pkg/apperror"
)

// ErrorHandler là nơi DUY NHẤT chuyển error thành HTTP response.
// Controller/middleware khác chỉ cần `c.Error(err)` rồi return.
//
//   - *apperror.AppError → dùng Status + Errors của nó.
//   - error bất kỳ khác  → 500, log chi tiết, client chỉ thấy thông báo chung.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err
		var appErr *apperror.AppError
		if !errors.As(err, &appErr) {
			appErr = apperror.Internal(err)
		}

		if appErr.Status >= http.StatusInternalServerError {
			slog.Error("request failed",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"error", appErr.Error(),
			)
		}

		c.AbortWithStatusJSON(appErr.Status, dto.NewErrorResponse(appErr.Errors))
	}
}

// Recovery bắt panic và trả về 500 theo đúng format lỗi RealWorld
// (gin.Recovery mặc định trả về body rỗng). Stack trace vẫn được Gin log ra.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			dto.NewErrorResponse(apperror.Internal(nil).Errors))
	})
}

// NotFound xử lý route không tồn tại.
func NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound,
			dto.NewErrorResponse(apperror.NotFound("route not found").Errors))
	}
}
