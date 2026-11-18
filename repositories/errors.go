package repositories

import (
	"errors"

	"gorm.io/gorm"
)

// Lỗi chung của tầng repository. Service chỉ so sánh với các lỗi này,
// nhờ vậy service không phụ thuộc trực tiếp vào GORM.
var (
	ErrNotFound  = errors.New("record not found")
	ErrDuplicate = errors.New("duplicate record")
)

func translateError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrDuplicate
	default:
		return err
	}
}
