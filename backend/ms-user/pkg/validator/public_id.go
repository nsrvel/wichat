package validator

import (
	"regexp"
	"strings"

	"github.com/wichat/wichat/backend/wi-shared/exception"
)

var groupPublicIDPattern = regexp.MustCompile(`^[0-9a-f]{24}$`)

// GroupPublicID normalizes and validates a client-facing group id.
func GroupPublicID(raw string) (string, error) {
	id := strings.TrimSpace(strings.ToLower(raw))
	if !groupPublicIDPattern.MatchString(id) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
	}

	return id, nil
}
