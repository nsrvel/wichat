package validator

import (
	"strings"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/password"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

// NewAccount validates email, password, and display name for registration.
func NewAccount(email, pass, displayName string) (normalizedEmail, name string, err error) {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return "", "", err
	}
	if err := password.Validate(pass); err != nil {
		return "", "", err
	}

	name = strings.TrimSpace(displayName)
	if name == "" || utf8.RuneCountInString(name) > constants.MaxDisplayNameRunes {
		return "", "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "display_name"})
	}

	return normalized, name, nil
}
