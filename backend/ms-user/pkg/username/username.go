package username

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

// Validation reason codes returned by CheckUsername.
const (
	ReasonTooShort      = "too_short"
	ReasonTooLong       = "too_long"
	ReasonInvalidFormat = "invalid_format"
	ReasonReserved      = "reserved"
	ReasonTaken         = "taken"
)

var reservedUsernames = map[string]struct{}{
	"admin":         {},
	"administrator": {},
	"api":           {},
	"auth":          {},
	"bot":           {},
	"channel":       {},
	"everyone":      {},
	"group":         {},
	"help":          {},
	"here":          {},
	"invite":        {},
	"login":         {},
	"mod":           {},
	"moderator":     {},
	"null":          {},
	"owner":         {},
	"register":      {},
	"root":          {},
	"settings":      {},
	"support":       {},
	"system":        {},
	"undefined":     {},
	"wichat":        {},
}

// Lookup resolves usernames for uniqueness checks.
type Lookup interface {
	FindUserByUsername(ctx context.Context, username string) (models.User, error)
}

// Validate normalizes a raw handle and checks length, ASCII format, separators, and reserved names.
func Validate(raw string) (normalized, reason string, err error) {
	u := strings.ToLower(strings.TrimSpace(raw))
	n := utf8.RuneCountInString(u)
	if n < MinLen {
		return u, ReasonTooShort, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "username", "reason": ReasonTooShort})
	}
	if n > MaxLen {
		return u, ReasonTooLong, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "username", "reason": ReasonTooLong})
	}
	if !isAlnumASCII(u[0]) || !isAlnumASCII(u[len(u)-1]) {
		return u, ReasonInvalidFormat, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "username", "reason": ReasonInvalidFormat})
	}

	var prevSep bool
	for i := 0; i < len(u); i++ {
		c := u[i]
		if c == '.' || c == '_' {
			if prevSep {
				return u, ReasonInvalidFormat, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "username", "reason": ReasonInvalidFormat})
			}
			prevSep = true
			continue
		}
		if !isAlnumASCII(c) {
			return u, ReasonInvalidFormat, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "username", "reason": ReasonInvalidFormat})
		}
		prevSep = false
	}

	if _, reserved := reservedUsernames[u]; reserved {
		return u, ReasonReserved, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "username", "reason": ReasonReserved})
	}

	return u, "", nil
}

// Normalize validates and lowercases a handle.
func Normalize(raw string) (string, error) {
	u, _, err := Validate(raw)

	return u, err
}

// Allocate picks a unique handle from the email local-part.
func Allocate(ctx context.Context, lookup Lookup, email string) (string, error) {
	base, err := BaseFromEmail(email)
	if err != nil {
		return "", err
	}

	for attempt := 0; attempt < 1000; attempt++ {
		cand := candidate(base, attempt)
		normalized, err := Normalize(cand)
		if err != nil {
			continue
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
		return fallbackBase, nil
	}

	normalized, err := Normalize(sanitized)
	if err != nil {
		return fallbackBase, nil
	}

	return normalized, nil
}

func isAlnumASCII(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
}

func candidate(base string, attempt int) string {
	if attempt == 0 {
		return base
	}

	suffix := fmt.Sprintf("%d", attempt+1)
	maxBase := MaxLen - len(suffix)
	trimmed := strings.TrimRight(truncateBytes(base, maxBase), "._")
	if trimmed == "" {
		trimmed = fallbackBase
	}

	return trimmed + suffix
}

func truncateBytes(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return s
	}

	return s[:maxBytes]
}

func sanitizeLocalPart(local string) string {
	local = strings.ToLower(strings.TrimSpace(local))
	var b strings.Builder
	var prevSep bool
	for i := 0; i < len(local); i++ {
		c := local[i]
		if c == '_' || c == '.' {
			if b.Len() > 0 && !prevSep {
				b.WriteByte(c)
				prevSep = true
			}
			continue
		}
		if isAlnumASCII(c) {
			b.WriteByte(c)
			prevSep = false
		}
	}

	return strings.Trim(b.String(), "._")
}
