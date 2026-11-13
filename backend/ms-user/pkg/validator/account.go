package validator

import (
	"strings"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/password"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

// DisplayName trims and validates a user display name.
func DisplayName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(name)
	if n < constants.MinDisplayNameRunes || n > constants.MaxDisplayNameRunes ||
		hasControlOrInvisible(name, false) || hasSpoofingChars(name) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "display_name"})
	}

	return name, nil
}

// NewAccount validates email, password, and display name for registration.
func NewAccount(email, pass, displayName string) (normalizedEmail, name string, err error) {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return "", "", err
	}
	if err := password.Validate(pass); err != nil {
		return "", "", err
	}

	name, err = DisplayName(displayName)
	if err != nil {
		return "", "", err
	}

	return normalized, name, nil
}
