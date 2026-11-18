package services

import (
	"context"
	"errors"
	"strings"

	"conduit/dto"
	"conduit/models"
	"conduit/pkg/apperror"
	"conduit/pkg/password"
	"conduit/repositories"
)

// TokenIssuer tách service khỏi implementation JWT cụ thể (dễ mock khi viết unit test).
type TokenIssuer interface {
	Generate(userID uint) (string, error)
}

type UserService interface {
	Register(ctx context.Context, in dto.RegisterUserInput) (*dto.UserResponse, error)
	Login(ctx context.Context, in dto.LoginUserInput) (*dto.UserResponse, error)
	// token là JWT hiện tại của request; được trả lại nguyên vẹn trong response.
	GetCurrentUser(ctx context.Context, userID uint, token string) (*dto.UserResponse, error)
	UpdateCurrentUser(ctx context.Context, userID uint, token string, in dto.UpdateUserInput) (*dto.UserResponse, error)
}

type userService struct {
	users  repositories.UserRepository
	tokens TokenIssuer
}

func NewUserService(users repositories.UserRepository, tokens TokenIssuer) UserService {
	return &userService{users: users, tokens: tokens}
}

func (s *userService) Register(ctx context.Context, in dto.RegisterUserInput) (*dto.UserResponse, error) {
	username := strings.TrimSpace(in.Username)
	email := normalizeEmail(in.Email)
	if username == "" {
		return nil, apperror.Unprocessable("username can't be blank")
	}

	if err := s.ensureUnique(ctx, 0, &email, &username); err != nil {
		return nil, err
	}

	hashed, err := password.Hash(in.Password)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	user := &models.User{Username: username, Email: email, PasswordHash: hashed}
	if err := s.users.Create(ctx, user); err != nil {
		// Trường hợp race condition: 2 request cùng đăng ký một email cùng lúc,
		// cả hai đều qua được ensureUnique, nhưng unique index của DB vẫn chặn được.
		if errors.Is(err, repositories.ErrDuplicate) {
			return nil, apperror.Unprocessable("email or username has already been taken")
		}
		return nil, err
	}

	return s.withNewToken(user)
}

func (s *userService) Login(ctx context.Context, in dto.LoginUserInput) (*dto.UserResponse, error) {
	user, err := s.users.FindByEmail(ctx, normalizeEmail(in.Email))
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, errInvalidCredentials()
		}
		return nil, err
	}
	if !password.Verify(user.PasswordHash, in.Password) {
		return nil, errInvalidCredentials()
	}
	return s.withNewToken(user)
}

func (s *userService) GetCurrentUser(ctx context.Context, userID uint, token string) (*dto.UserResponse, error) {
	user, err := s.findCurrentUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewUserResponse(user, token), nil
}

func (s *userService) UpdateCurrentUser(ctx context.Context, userID uint, token string, in dto.UpdateUserInput) (*dto.UserResponse, error) {
	user, err := s.findCurrentUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Gom tất cả lỗi validate để trả về một lần, client sửa được hết trong một lượt.
	var problems []string
	var newEmail, newUsername *string

	if in.Email != nil {
		email := normalizeEmail(*in.Email)
		switch {
		case email == "":
			problems = append(problems, "email can't be blank")
		case email != user.Email:
			newEmail = &email
		}
	}
	if in.Username != nil {
		username := strings.TrimSpace(*in.Username)
		switch {
		case username == "":
			problems = append(problems, "username can't be blank")
		case username != user.Username:
			newUsername = &username
		}
	}
	if in.Password != nil && *in.Password == "" {
		problems = append(problems, "password can't be blank")
	}
	if len(problems) > 0 {
		return nil, apperror.Unprocessable(problems...)
	}

	if err := s.ensureUnique(ctx, user.ID, newEmail, newUsername); err != nil {
		return nil, err
	}

	if newEmail != nil {
		user.Email = *newEmail
	}
	if newUsername != nil {
		user.Username = *newUsername
	}
	if in.Password != nil {
		hashed, err := password.Hash(*in.Password)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		user.PasswordHash = hashed
	}
	if in.Bio != nil {
		user.Bio = nilIfBlank(*in.Bio)
	}
	if in.Image != nil {
		user.Image = nilIfBlank(*in.Image)
	}

	if err := s.users.Update(ctx, user); err != nil {
		if errors.Is(err, repositories.ErrDuplicate) {
			return nil, apperror.Unprocessable("email or username has already been taken")
		}
		return nil, err
	}

	// Token cũ vẫn hợp lệ vì JWT chỉ chứa user ID (không đổi khi update).
	return dto.NewUserResponse(user, token), nil
}

// ==================== helpers ====================

func (s *userService) findCurrentUser(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			// Token hợp lệ nhưng user đã bị xoá → coi như chưa xác thực.
			return nil, apperror.Unauthorized("user no longer exists")
		}
		return nil, err
	}
	return user, nil
}

// ensureUnique kiểm tra email/username (nil = bỏ qua) chưa bị user khác sử dụng.
func (s *userService) ensureUnique(ctx context.Context, excludeID uint, email, username *string) error {
	var problems []string
	if email != nil {
		taken, err := s.users.ExistsByEmail(ctx, *email, excludeID)
		if err != nil {
			return err
		}
		if taken {
			problems = append(problems, "email has already been taken")
		}
	}
	if username != nil {
		taken, err := s.users.ExistsByUsername(ctx, *username, excludeID)
		if err != nil {
			return err
		}
		if taken {
			problems = append(problems, "username has already been taken")
		}
	}
	if len(problems) > 0 {
		return apperror.Unprocessable(problems...)
	}
	return nil
}

func (s *userService) withNewToken(user *models.User) (*dto.UserResponse, error) {
	token, err := s.tokens.Generate(user.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return dto.NewUserResponse(user, token), nil
}

// Không tiết lộ là sai email hay sai mật khẩu để tránh dò tìm tài khoản.
func errInvalidCredentials() *apperror.AppError {
	return apperror.Unauthorized("email or password is invalid")
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func nilIfBlank(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
