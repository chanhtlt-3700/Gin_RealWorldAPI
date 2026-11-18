package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"conduit/dto"
	"conduit/middlewares"
	"conduit/services"
)

// UserController chỉ lo phần HTTP: parse request → gọi service → ghi response.
// Không chứa business logic.
type UserController struct {
	users services.UserService
}

func NewUserController(users services.UserService) *UserController {
	return &UserController{users: users}
}

// Register godoc
// POST /api/users — 201 Created
func (h *UserController) Register(c *gin.Context) {
	var req dto.RegisterUserRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.users.Register(c.Request.Context(), req.User)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

// Login godoc
// POST /api/users/login — 200 OK
func (h *UserController) Login(c *gin.Context) {
	var req dto.LoginUserRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.users.Login(c.Request.Context(), req.User)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetCurrent godoc
// GET /api/user — 200 OK (Auth required)
func (h *UserController) GetCurrent(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	res, err := h.users.GetCurrentUser(c.Request.Context(), userID, middlewares.CurrentToken(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// Update godoc
// PUT /api/user — 200 OK (Auth required)
func (h *UserController) Update(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	var req dto.UpdateUserRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.users.UpdateCurrentUser(c.Request.Context(), userID, middlewares.CurrentToken(c), req.User)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, res)
}
