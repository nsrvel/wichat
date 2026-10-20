package validator

import (
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

// InstanceSetup validates the owner setup wizard payload.
func InstanceSetup(in models.CompleteInstanceSetupInput) (models.InstanceSetupPayload, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" || utf8.RuneCountInString(name) > constants.MaxGroupNameRunes {
		return models.InstanceSetupPayload{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"})
	}

	logoURL := strings.TrimSpace(in.LogoURL)
	if logoURL != "" {
		if utf8.RuneCountInString(logoURL) > constants.MaxLogoURLRunes {
			return models.InstanceSetupPayload{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "logo_url"})
		}
		u, err := url.Parse(logoURL)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return models.InstanceSetupPayload{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "logo_url"})
		}
	}

	mode := strings.TrimSpace(in.RegistrationMode)
	switch mode {
	case constants.RegistrationInviteOnly, constants.RegistrationApprovalRequired, constants.RegistrationOpen:
	default:
		return models.InstanceSetupPayload{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "registration_mode"})
	}

	return models.InstanceSetupPayload{
		GroupName:        name,
		GroupLogoURL:     logoURL,
		UsernameEnabled:  in.UsernameEnabled,
		RegistrationMode: mode,
		GuestEnabled:     in.GuestEnabled,
	}, nil
}
