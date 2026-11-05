package username

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

const (
	MinLen       = 3
	MaxLen       = 30
	fallbackBase = "user"
)

// Lookup resolves usernames for uniqueness checks.
type Lookup interface {
	FindUserByUsername(ctx context.Context, username string) (models.User, error)
}

// Normalize validates and lowercases a handle.
func Normalize(raw string) (string, error) {
	username := strings.ToLower(strings.TrimSpace(raw))
	if utf8.RuneCountInString(username) < MinLen || utf8.RuneCountInString(username) > MaxLen {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "username"})
	}
	for _, r := range username {
		if r == '_' || r == '.' {
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "username"})
		}
	}

	return username, nil
}

// Allocate picks a unique handle from the email local-part.
func Allocate(ctx context.Context, lookup Lookup, email string) (string, error) {
	base, err := BaseFromEmail(email)
	if err != nil {
		return "", err
	}

	for attempt := 0; attempt < 1000; attempt++ {
		candidate := candidate(base, attempt)
		normalized, err := Normalize(candidate)
		if err != nil {
			return "", err
		}

		_, err = lookup.FindUserByUsername(ctx, normalized)
		if errors.Is(err, constants.ErrNotFound) {
			return normalized, nil
		}
		if err != nil {
			return "", err
		}
	}

	return "", exception.Internal(fmt.Errorf("username allocation exhausted for %q", email))
}

// BaseFromEmail derives a handle base from an email address.
func BaseFromEmail(email string) (string, error) {
	local := strings.TrimSpace(email)
	if at := strings.LastIndex(local, "@"); at >= 0 {
		local = local[:at]
	}
	if plus := strings.Index(local, "+"); plus >= 0 {
		local = local[:plus]
	}

	sanitized := sanitizeLocalPart(local)
	if sanitized == "" {
		sanitized = fallbackBase
	}

	normalized, err := Normalize(sanitized)
	if err != nil {
		if utf8.RuneCountInString(sanitized) < MinLen {
			return fallbackBase, nil
		}

		return "", err
	}

	return normalized, nil
}

func candidate(base string, attempt int) string {
	if attempt == 0 {
		return base
	}

	suffix := fmt.Sprintf("%d", attempt+1)
	maxBase := MaxLen - utf8.RuneCountInString(suffix)
	trimmed := truncateRunes(base, maxBase)

	return trimmed + suffix
}

func truncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}

	var n int
	for i := range s {
		if n == maxRunes {
			return s[:i]
		}
		n++
	}

	return s
}

func sanitizeLocalPart(local string) string {
	local = strings.ToLower(strings.TrimSpace(local))
	var b strings.Builder
	for _, r := range local {
		if r == '_' || r == '.' {
			b.WriteRune(r)
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}

	return strings.Trim(b.String(), "._")
}
