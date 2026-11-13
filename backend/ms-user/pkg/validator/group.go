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
	n := utf8.RuneCountInString(trimmed)
	if n < constants.MinGroupNameRunes || n > constants.MaxGroupNameRunes || hasControlOrInvisible(trimmed, false) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"})
	}

	return trimmed, nil
}

// GroupDescription trims and validates group profile description.
func GroupDescription(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if utf8.RuneCountInString(trimmed) > constants.MaxGroupDescriptionRunes || hasControlOrInvisible(trimmed, true) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "description"})
	}

	return trimmed, nil
}

// GroupBannerColor normalizes a solid banner palette id.
func GroupBannerColor(raw string) (string, error) {
	id := strings.TrimSpace(strings.ToLower(raw))
	if id == "" || id == "default" {
		return "", nil
	}
	if _, ok := constants.AllowedGroupBannerColors[id]; !ok {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "banner_color"})
	}

	return id, nil
}

func validateGroupImageUpload(contentType string, size, maxBytes int64) error {
	if size <= 0 || size > maxBytes {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"})
	}
	switch strings.TrimSpace(strings.ToLower(contentType)) {
	case "image/png", "image/jpeg", "image/webp", "image/gif":
		return nil
	default:
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"})
	}
}

// GroupIconUpload validates group icon upload metadata.
func GroupIconUpload(contentType string, size int64) error {
	return validateGroupImageUpload(contentType, size, constants.MaxGroupIconBytes)
}

// GroupBannerUpload validates group banner image upload metadata.
func GroupBannerUpload(contentType string, size int64) error {
	return validateGroupImageUpload(contentType, size, constants.MaxGroupBannerBytes)
}
