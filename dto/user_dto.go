package dto

import "conduit/models"

// ==================== Requests ====================

// RegisterUserRequest: POST /api/users
//
//	{"user": {"username": "jake", "email": "jake@jake.jake", "password": "jakejake"}}
type RegisterUserRequest struct {
	User RegisterUserInput `json:"user"`
}

type RegisterUserInput struct {
	Username string `json:"username" binding:"required,max=50"`
	Email    string `json:"email" binding:"required,email,max=255"`
	// bcrypt chỉ dùng 72 byte đầu tiên, nên giới hạn max=72 để tránh hiểu nhầm.
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// LoginUserRequest: POST /api/users/login
//
//	{"user": {"email": "jake@jake.jake", "password": "jakejake"}}
type LoginUserRequest struct {
	User LoginUserInput `json:"user"`
}

type LoginUserInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateUserRequest: PUT /api/user
//
// Mọi field đều là con trỏ để phân biệt "không gửi" (nil → giữ nguyên)
// với "gửi giá trị" (non-nil → cập nhật). Gửi chuỗi rỗng cho bio/image sẽ xoá giá trị.
type UpdateUserRequest struct {
	User UpdateUserInput `json:"user"`
}

type UpdateUserInput struct {
	Email    *string `json:"email" binding:"omitempty,email,max=255"`
	Username *string `json:"username" binding:"omitempty,max=50"`
	Password *string `json:"password" binding:"omitempty,min=8,max=72"`
	Bio      *string `json:"bio"`
	Image    *string `json:"image" binding:"omitempty,max=2048"`
}

// ==================== Responses ====================

// UserResponse:
//
//	{"user": {"email": "...", "token": "...", "username": "...", "bio": null, "image": null}}
type UserResponse struct {
	User UserData `json:"user"`
}

type UserData struct {
	Email    string  `json:"email"`
	Token    string  `json:"token"`
	Username string  `json:"username"`
	Bio      *string `json:"bio"`
	Image    *string `json:"image"`
}

func NewUserResponse(u *models.User, token string) *UserResponse {
	return &UserResponse{
		User: UserData{
			Email:    u.Email,
			Token:    token,
			Username: u.Username,
			Bio:      u.Bio,
			Image:    u.Image,
		},
	}
}
