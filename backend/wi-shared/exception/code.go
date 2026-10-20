package exception

import "strings"

// MessageCode is a stable i18n key emitted to clients. Each const must exist in the
// frontend translation catalog (and go-i18n for server-rendered text when applicable).
type MessageCode string

const (
	MsgInternal     MessageCode = "error.internal"
	MsgInvalidInput MessageCode = "error.invalid_input"
	MsgUnavailable  MessageCode = "error.unavailable"
	MsgTimeout      MessageCode = "error.timeout"
	MsgRateLimited  MessageCode = "error.rate_limited"

	MsgUnauthorized          MessageCode = "auth.unauthorized"
	MsgInvalidCredentials    MessageCode = "auth.invalid_credentials"
	MsgSetupClosed           MessageCode = "auth.setup_closed"
	MsgRegistrationClosed    MessageCode = "auth.registration_closed"
	MsgGuestDisabled         MessageCode = "auth.guest_disabled"
	MsgInstanceSetupRequired MessageCode = "auth.instance_setup_required"
	MsgForbidden             MessageCode = "auth.forbidden"
	MsgEmailTaken            MessageCode = "auth.email_taken"
	MsgUsernameTaken         MessageCode = "auth.username_taken"
	MsgAgeRestricted         MessageCode = "auth.age_restricted"
	MsgGoogleDenied          MessageCode = "auth.google_denied"
	MsgGoogleEmailUnverified MessageCode = "auth.google_email_unverified"
	MsgResetCodeInvalid      MessageCode = "auth.reset_code_invalid"
	MsgResetRateLimited      MessageCode = "auth.reset_rate_limited"

	MsgUserNotFound MessageCode = "user.not_found"
)

// ValidMessageCode reports whether code is an allowed i18n key shape from trusted services.
func ValidMessageCode(code MessageCode) bool {
	s := string(code)
	if s == "" {
		return false
	}

	return strings.HasPrefix(s, "error.") ||
		strings.HasPrefix(s, "auth.") ||
		strings.HasPrefix(s, "user.")
}
