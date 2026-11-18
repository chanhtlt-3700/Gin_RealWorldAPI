package repositories

import (
	"context"

	"gorm.io/gorm"

	"conduit/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uint) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	// ExistsByEmail / ExistsByUsername bỏ qua user có ID = excludeID
	// (dùng khi update: không tính chính user đang sửa). Truyền 0 để không loại trừ.
	ExistsByEmail(ctx context.Context, email string, excludeID uint) (bool, error)
	ExistsByUsername(ctx context.Context, username string, excludeID uint) (bool, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return translateError(r.db.WithContext(ctx).Create(user).Error)
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return translateError(r.db.WithContext(ctx).Save(user).Error)
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.findOne(ctx, "email = ?", email)
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	return r.findOne(ctx, "username = ?", username)
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string, excludeID uint) (bool, error) {
	return r.exists(ctx, excludeID, "email = ?", email)
}

// So sánh không phân biệt hoa thường để "Jake" và "jake" không thể cùng tồn tại.
func (r *userRepository) ExistsByUsername(ctx context.Context, username string, excludeID uint) (bool, error) {
	return r.exists(ctx, excludeID, "LOWER(username) = LOWER(?)", username)
}

func (r *userRepository) findOne(ctx context.Context, query string, args ...any) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where(query, args...).First(&user).Error; err != nil {
		return nil, translateError(err)
	}
	return &user, nil
}

func (r *userRepository) exists(ctx context.Context, excludeID uint, query string, args ...any) (bool, error) {
	q := r.db.WithContext(ctx).Model(&models.User{}).Where(query, args...)
	if excludeID != 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, translateError(err)
	}
	return count > 0, nil
}
