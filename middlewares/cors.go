package middlewares

import (
	"net/http"
	"slices"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS cho phép frontend (React/Vue/Angular Conduit...) gọi API từ domain khác.
// allowedOrigins rỗng hoặc chứa "*" → cho phép mọi origin.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	cfg := cors.Config{
		AllowMethods: []string{
			http.MethodGet, http.MethodPost, http.MethodPut,
			http.MethodPatch, http.MethodDelete, http.MethodOptions,
		},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		MaxAge:       12 * time.Hour,
	}

	if len(allowedOrigins) == 0 || slices.Contains(allowedOrigins, "*") {
		cfg.AllowAllOrigins = true
	} else {
		cfg.AllowOrigins = allowedOrigins
	}
	return cors.New(cfg)
}
