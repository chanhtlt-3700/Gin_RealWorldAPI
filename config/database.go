package config

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite" // SQLite driver thuần Go (không cần CGO/gcc, chạy tốt trên Windows)
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"conduit/models"
)

// ConnectDatabase mở kết nối tới SQLite hoặc PostgreSQL tuỳ theo DB_DRIVER.
func ConnectDatabase(cfg DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case DriverPostgres:
		dialector = postgres.Open(cfg.PostgresDSN())
	default:
		dialector = sqlite.Open(cfg.SQLiteDSN())
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(parseLogLevel(cfg.LogLevel)),
		// Dịch lỗi riêng của từng DB (vd: unique violation) thành lỗi chung của GORM
		// như gorm.ErrDuplicatedKey, để tầng repository không phụ thuộc vào driver.
		TranslateError: true,
		NowFunc:        func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("database: open %s: %w", cfg.Driver, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: get sql.DB: %w", err)
	}
	if cfg.Driver == DriverPostgres {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(25)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database: ping %s: %w", cfg.Driver, err)
	}
	return db, nil
}

// MigrateDatabase tạo/cập nhật bảng cho tất cả models đã đăng ký trong models.All().
func MigrateDatabase(db *gorm.DB) error {
	if err := db.AutoMigrate(models.All()...); err != nil {
		return fmt.Errorf("database: auto migrate: %w", err)
	}
	return nil
}

func parseLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}
