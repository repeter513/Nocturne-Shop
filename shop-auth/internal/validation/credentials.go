// Package validation holds shared input checks for auth RPC handlers.
// Пакет validation — общие проверки входных данных для auth RPC.
package validation

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

var ErrInvalidInput = errors.New("invalid input")

const (
	maxEmailLen      = 50
	minPasswordRunes = 8
	maxPasswordBytes = 72
)

// Credentials normalizes email and validates email/password bounds for Register/Login.
// Credentials нормализует email и проверяет ограничения email/пароля для Register/Login.
func Credentials(email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || len(email) > maxEmailLen {
		return "", ErrInvalidInput
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "", ErrInvalidInput
	}
	if utf8.RuneCountInString(password) < minPasswordRunes {
		return "", ErrInvalidInput
	}
	if len(password) > maxPasswordBytes {
		return "", ErrInvalidInput
	}
	return email, nil
}
