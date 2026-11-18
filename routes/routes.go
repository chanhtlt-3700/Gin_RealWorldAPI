package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"conduit/config"
	"conduit/controllers"
	"conduit/middlewares"
)

// Handlers gom tất cả controller. Mỗi giai đoạn chỉ cần thêm field mới
// (Profile, Article, Comment, Tag...) và đăng ký route tương ứng.
type Handlers struct {
	User *controllers.UserController
}

func SetupRouter(cfg *config.Config, tokens middlewares.TokenParser, h Handlers) *gin.Engine {
	if cfg.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	_ = r.SetTrustedProxies(nil) // không tin header X-Forwarded-For từ client bất kỳ

	// Thứ tự quan trọng: Recovery bọc ngoài cùng để bắt panic ở mọi tầng bên trong;
	// ErrorHandler nằm trong cùng để xử lý lỗi do controller/auth middleware ghi lại.
	r.Use(
		gin.Logger(),
		middlewares.Recovery(),
		middlewares.CORS(cfg.CORS.AllowedOrigins),
		middlewares.ErrorHandler(),
	)
	r.NoRoute(middlewares.NotFound())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authRequired := middlewares.AuthRequired(tokens)

	api := r.Group("/api")
	registerUserRoutes(api, h.User, authRequired)

	return r
}

func registerUserRoutes(api *gin.RouterGroup, h *controllers.UserController, authRequired gin.HandlerFunc) {
	api.POST("/users", h.Register)
	api.POST("/users/login", h.Login)

	user := api.Group("/user", authRequired)
	user.GET("", h.GetCurrent)
	user.PUT("", h.Update)
}
