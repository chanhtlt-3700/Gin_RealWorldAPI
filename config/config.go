package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"

	DriverSQLite   = "sqlite"
	DriverPostgres = "postgres"
)

// Config gom toàn bộ cấu hình của ứng dụng, được load một lần lúc khởi động
// rồi truyền (inject) xuống các tầng cần dùng — không dùng biến global.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type AppConfig struct {
	Env  string
	Port string
}

func (a AppConfig) IsProduction() bool { return a.Env == EnvProduction }

type DatabaseConfig struct {
	Driver   string
	LogLevel string

	SQLitePath string

	URL      string // DATABASE_URL, ưu tiên hơn các field rời bên dưới
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
	Issuer    string
}

type CORSConfig struct {
	AllowedOrigins []string
}

// Load đọc file .env (nếu có) rồi đọc biến môi trường.
// Biến môi trường thật luôn được ưu tiên hơn giá trị trong .env.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("config: load .env: %w", err)
	}

	jwtTTL, err := time.ParseDuration(getEnv("JWT_EXPIRES_IN", "72h"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid JWT_EXPIRES_IN: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Env:  strings.ToLower(getEnv("APP_ENV", EnvDevelopment)),
			Port: getEnv("PORT", "8080"),
		},
		Database: DatabaseConfig{
			Driver:     strings.ToLower(getEnv("DB_DRIVER", DriverSQLite)),
			LogLevel:   strings.ToLower(getEnv("DB_LOG_LEVEL", "warn")),
			SQLitePath: getEnv("SQLITE_PATH", "conduit.db"),
			URL:        os.Getenv("DATABASE_URL"),
			Host:       getEnv("DB_HOST", "localhost"),
			Port:       getEnv("DB_PORT", "5432"),
			User:       getEnv("DB_USER", "postgres"),
			Password:   os.Getenv("DB_PASSWORD"),
			Name:       getEnv("DB_NAME", "conduit"),
			SSLMode:    getEnv("DB_SSLMODE", "disable"),
			TimeZone:   getEnv("DB_TIMEZONE", "UTC"),
		},
		JWT: JWTConfig{
			Secret:    os.Getenv("JWT_SECRET"),
			ExpiresIn: jwtTTL,
			Issuer:    getEnv("JWT_ISSUER", "conduit"),
		},
		CORS: CORSConfig{
			AllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "*")),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate fail-fast ngay lúc khởi động thay vì để lỗi cấu hình lộ ra lúc runtime.
func (c *Config) validate() error {
	var errs []error

	switch {
	case c.JWT.Secret == "":
		errs = append(errs, errors.New("JWT_SECRET is required"))
	case c.App.IsProduction() && len(c.JWT.Secret) < 32:
		errs = append(errs, errors.New("JWT_SECRET must be at least 32 characters in production"))
	}
	if c.JWT.ExpiresIn <= 0 {
		errs = append(errs, errors.New("JWT_EXPIRES_IN must be positive"))
	}
	if c.Database.Driver != DriverSQLite && c.Database.Driver != DriverPostgres {
		errs = append(errs, fmt.Errorf("DB_DRIVER must be %q or %q, got %q", DriverSQLite, DriverPostgres, c.Database.Driver))
	}

	if len(errs) > 0 {
		return fmt.Errorf("config: %w", errors.Join(errs...))
	}
	return nil
}

// PostgresDSN trả về connection string dạng URL; dùng url.URL để escape
// đúng các ký tự đặc biệt trong user/password.
func (d DatabaseConfig) PostgresDSN() string {
	if d.URL != "" {
		return d.URL
	}
	q := url.Values{}
	q.Set("sslmode", d.SSLMode)
	q.Set("TimeZone", d.TimeZone)

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(d.User, d.Password),
		Host:     net.JoinHostPort(d.Host, d.Port),
		Path:     "/" + d.Name,
		RawQuery: q.Encode(),
	}
	return u.String()
}

// SQLiteDSN bật foreign keys (SQLite tắt mặc định), WAL mode và busy timeout
// để tránh lỗi "database is locked" khi có nhiều request ghi đồng thời.
func (d DatabaseConfig) SQLiteDSN() string {
	return d.SQLitePath + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
