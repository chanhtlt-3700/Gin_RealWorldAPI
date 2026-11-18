// Package password băm và kiểm tra mật khẩu bằng bcrypt.
package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func Hash(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("password: hash: %w", err)
	}
	return string(hashed), nil
}

// Verify so sánh constant-time giữa hash đã lưu và mật khẩu người dùng nhập.
func Verify(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
