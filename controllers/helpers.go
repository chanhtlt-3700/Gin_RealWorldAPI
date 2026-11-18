package controllers

import (
	"github.com/gin-gonic/gin"

	"conduit/middlewares"
	"conduit/pkg/apperror"
	"conduit/pkg/validation"
)

// bindJSON parse + validate body. Trả về false (và đã ghi lỗi 422 vào context) nếu không hợp lệ.
func bindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		_ = c.Error(validation.Translate(err))
		return false
	}
	return true
}

// requireUser lấy user ID đã được AuthRequired gắn vào context.
// Kiểm tra phòng thủ: nếu route quên gắn middleware thì trả 401 thay vì xử lý sai.
func requireUser(c *gin.Context) (uint, bool) {
	userID, ok := middlewares.CurrentUserID(c)
	if !ok {
		_ = c.Error(apperror.Unauthorized())
		return 0, false
	}
	return userID, true
}
