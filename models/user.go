package models

import "time"

// User là entity lưu trong DB. Model không bao giờ được trả thẳng ra JSON —
// việc format response là trách nhiệm của package dto.
//
// Không dùng gorm.Model vì soft delete (DeletedAt) sẽ xung đột với unique index
// của username/email khi người dùng đăng ký lại.
type User struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"size:50;not null;uniqueIndex"`
	Email        string    `gorm:"size:255;not null;uniqueIndex"`
	PasswordHash string    `gorm:"size:255;not null"`
	Bio          *string   `gorm:"type:text"` // nullable: spec trả về null khi chưa có
	Image        *string   `gorm:"type:text"` // nullable
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}
