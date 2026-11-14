package validator

import (
	"strings"
	"unicode/utf8"

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
