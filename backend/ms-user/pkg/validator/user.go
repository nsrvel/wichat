package validator

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

// UserAvatarUpload validates user avatar upload metadata.
func UserAvatarUpload(contentType string, size int64) error {
	return validateGroupImageUpload(contentType, size, constants.MaxUserAvatarBytes)
}

// UserProfileInput validates and normalizes user-editable profile fields.
func UserProfileInput(displayName, pronouns, bio, bannerColor string) (string, string, string, string, error) {
	dn, err := DisplayName(displayName)
	if err != nil {
		return "", "", "", "", err
	}

	p := strings.TrimSpace(pronouns)
	if utf8.RuneCountInString(p) > constants.MaxUserPronounsRunes {
		return "", "", "", "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "pronouns"})
	}

	b := strings.TrimSpace(bio)
	if utf8.RuneCountInString(b) > constants.MaxUserBioRunes {
		return "", "", "", "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "bio"})
	}

	c := strings.TrimSpace(bannerColor)
	if c != "" {
		if !hexColorPattern.MatchString(c) {
			return "", "", "", "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "banner_color"})
		}
		c = strings.ToUpper(c)
	}

	return dn, p, b, c, nil
}

// UserStatusNoteInput validates and normalizes user status note text, emoji, and optional expiry.
func UserStatusNoteInput(text, emoji, expiresAtRaw string, now time.Time) (models.UserStatusNotePatch, error) {
	t := strings.TrimSpace(text)
	if utf8.RuneCountInString(t) > constants.MaxUserStatusNoteRunes {
		return models.UserStatusNotePatch{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "text"})
	}

	e := strings.TrimSpace(emoji)
	if utf8.RuneCountInString(e) > constants.MaxUserStatusEmojiRunes {
		return models.UserStatusNotePatch{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "emoji"})
	}

	if t == "" && e == "" {
		return models.UserStatusNotePatch{}, nil
	}

	expRaw := strings.TrimSpace(expiresAtRaw)
	if expRaw == "" {
		return models.UserStatusNotePatch{Text: t, Emoji: e}, nil
	}

	parsed, err := time.Parse(time.RFC3339Nano, expRaw)
	if err != nil {
		return models.UserStatusNotePatch{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "expires_at"})
	}
	utc := parsed.UTC()
	if !utc.After(now) || utc.Sub(now) > constants.MaxUserStatusNoteDuration {
		return models.UserStatusNotePatch{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "expires_at"})
	}

	return models.UserStatusNotePatch{Text: t, Emoji: e, ExpiresAt: &utc}, nil
}

// UserPreferencesInput validates and normalizes user display locale and time format preferences.
func UserPreferencesInput(locale, timeFormat string) (models.UserPreferencesPatch, error) {
	loc := strings.ToLower(strings.TrimSpace(locale))
	switch loc {
	case "en", "id":
	default:
		return models.UserPreferencesPatch{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "locale"})
	}

	tf := strings.ToLower(strings.TrimSpace(timeFormat))
	switch tf {
	case "auto", "12h", "24h":
	default:
		return models.UserPreferencesPatch{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "time_format"})
	}

	return models.UserPreferencesPatch{Locale: loc, TimeFormat: tf}, nil
}
