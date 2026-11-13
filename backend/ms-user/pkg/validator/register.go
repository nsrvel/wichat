package validator

import (
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/username"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

// RegisterInput validates signup fields including username, birth date, and minimum age.
func RegisterInput(in models.RegisterInput) (email, handle, displayName string, birth *time.Time, err error) {
	email, displayName, err = NewAccount(in.Email, in.Password, in.DisplayName)
	if err != nil {
		return "", "", "", nil, err
	}

	handle, err = username.Normalize(in.Username)
	if err != nil {
		return "", "", "", nil, err
	}

	birth, err = ParseBirthDate(in.BirthDate)
	if err != nil {
		return "", "", "", nil, err
	}
	if !MeetsMinAge(*birth) {
		return "", "", "", nil, exception.InvalidInput(exception.MsgAgeRestricted, nil)
	}

	return email, handle, displayName, birth, nil
}
