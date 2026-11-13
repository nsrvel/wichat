package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

var hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// RoleName trims and validates a group role name.
func RoleName(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(trimmed)
	if n < constants.MinRoleNameRunes || n > constants.MaxRoleNameRunes || hasControlOrInvisible(trimmed, false) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"})
	}

	return trimmed, nil
}

// RoleColor validates an optional 6-digit hex color string.
func RoleColor(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !hexColorPattern.MatchString(trimmed) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "color"})
	}

	return trimmed, nil
}
