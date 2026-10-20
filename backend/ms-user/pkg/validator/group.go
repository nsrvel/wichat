package validator

import (
	"strings"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

// GroupName trims and validates a group display name.
func GroupName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > constants.MaxGroupNameRunes {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"})
	}

	return trimmed, nil
}
