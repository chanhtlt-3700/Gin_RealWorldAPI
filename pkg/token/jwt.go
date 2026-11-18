// Package token phát hành và xác thực JWT (HS256).
package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

func NewManager(secret string, ttl time.Duration, issuer string) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl, issuer: issuer}
}

// Generate tạo JWT cho userID, hết hạn sau ttl.
func (m *Manager) Generate(userID uint) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("token: sign: %w", err)
	}
	return signed, nil
}

// Parse xác thực chữ ký, thuật toán, issuer, thời hạn và trả về claims.
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		// Chặn tấn công "alg confusion" (vd: alg=none) bằng cách chỉ chấp nhận HS256.
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if claims.UserID == 0 {
		return nil, fmt.Errorf("%w: missing user_id claim", ErrInvalidToken)
	}
	return claims, nil
}
