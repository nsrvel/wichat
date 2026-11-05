package validator

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

// GuestRegister validates guest signup display name and birth date.
func GuestRegister(in models.GuestRegisterInput) (displayName string, birth *time.Time, err error) {
	name := strings.TrimSpace(in.DisplayName)
	if name == "" || utf8.RuneCountInString(name) > constants.MaxDisplayNameRunes {
		return "", nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "display_name"})
	}

	birth, err = ParseBirthDate(in.BirthDate)
	if err != nil {
		return "", nil, err
	}
	if !MeetsMinAge(*birth) {
		return "", nil, exception.InvalidInput(exception.MsgAgeRestricted, nil)
	}

	return name, birth, nil
}
