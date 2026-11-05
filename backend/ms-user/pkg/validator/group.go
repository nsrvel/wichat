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

// GroupIconUpload validates group icon upload metadata.
func GroupIconUpload(contentType string, size int64) error {
	if size <= 0 || size > constants.MaxGroupIconBytes {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"})
	}
	switch strings.TrimSpace(strings.ToLower(contentType)) {
	case "image/png", "image/jpeg", "image/webp":
		return nil
	default:
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"})
	}
}
