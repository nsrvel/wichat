package validator

import (
	"strings"

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
