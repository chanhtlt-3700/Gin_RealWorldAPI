package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"

	"conduit/pkg/apperror"
	"conduit/pkg/token"
)

const (
	ctxUserIDKey = "auth.userID"
	ctxTokenKey  = "auth.token"
)

// TokenParser: middleware chỉ cần khả năng parse token, không cần biết JWT cụ thể.
type TokenParser interface {
	Parse(tokenString string) (*token.Claims, error)
}

// AuthRequired bắt buộc header "Authorization: Token <jwt>", ngược lại trả 401.
func AuthRequired(parser TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := extractToken(c.GetHeader("Authorization"))
		if !ok {
			abortUnauthorized(c, "missing or malformed authorization header")
			return
		}
		claims, err := parser.Parse(raw)
		if err != nil {
			abortUnauthorized(c, "invalid or expired token")
			return
		}
		c.Set(ctxUserIDKey, claims.UserID)
		c.Set(ctxTokenKey, raw)
		c.Next()
	}
}

// AuthOptional dùng cho endpoint public nhưng trả thêm dữ liệu khi đã đăng nhập
// (vd: field "following", "favorited"). Token sai/hết hạn được coi như khách,
// để client giữ token cũ vẫn xem được nội dung công khai.
func AuthOptional(parser TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		if raw, ok := extractToken(c.GetHeader("Authorization")); ok {
			if claims, err := parser.Parse(raw); err == nil {
				c.Set(ctxUserIDKey, claims.UserID)
				c.Set(ctxTokenKey, raw)
			}
		}
		c.Next()
	}
}

// CurrentUserID trả về ID của user đã xác thực (ok=false nếu là khách).
func CurrentUserID(c *gin.Context) (uint, bool) {
	id, ok := c.Get(ctxUserIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := id.(uint)
	return userID, ok
}

// CurrentToken trả về JWT thô của request hiện tại.
func CurrentToken(c *gin.Context) string {
	return c.GetString(ctxTokenKey)
}

// extractToken chấp nhận "Token <jwt>" (chuẩn RealWorld) và "Bearer <jwt>" (tiện test).
func extractToken(header string) (string, bool) {
	scheme, raw, found := strings.Cut(strings.TrimSpace(header), " ")
	if !found {
		return "", false
	}
	if !strings.EqualFold(scheme, "Token") && !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	raw = strings.TrimSpace(raw)
	return raw, raw != ""
}

func abortUnauthorized(c *gin.Context, msg string) {
	_ = c.Error(apperror.Unauthorized(msg))
	c.Abort()
}
