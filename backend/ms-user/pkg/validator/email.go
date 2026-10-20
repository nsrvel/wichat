package validator

import (
	"strings"

	"github.com/wichat/wichat/backend/wi-shared/exception"
)

// NormalizeEmail lowercases, trims, and validates a signup/login email.
func NormalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if !validEmail(normalized) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "email"})
	}

	return normalized, nil
}

func validEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 || strings.ContainsAny(email, " \t\r\n") {
		return false
	}

	at := strings.IndexByte(email, '@')
	if at <= 0 || at != strings.LastIndexByte(email, '@') || at >= len(email)-1 {
		return false
	}

	return strings.Contains(email[at+1:], ".")
}
