package validator

import (
	"strings"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

// ChannelPublicID normalizes and validates a client-facing channel id.
func ChannelPublicID(raw string) (string, error) {
	id := strings.TrimSpace(strings.ToLower(raw))
	if !groupPublicIDPattern.MatchString(id) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channel_id"})
	}

	return id, nil
}

// ChannelName trims and validates a channel name.
func ChannelName(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(trimmed)
	if n < constants.MinChannelNameRunes || n > constants.MaxChannelNameRunes || hasControlOrInvisible(trimmed, false) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"})
	}

	return trimmed, nil
}
