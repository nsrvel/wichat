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
	MsgInstanceSetupRequired MessageCode = "auth.instance_setup_required"
	MsgForbidden             MessageCode = "auth.forbidden"
	MsgEmailTaken            MessageCode = "auth.email_taken"
	MsgUsernameTaken         MessageCode = "auth.username_taken"
	MsgAgeRestricted         MessageCode = "auth.age_restricted"
	MsgGoogleDenied          MessageCode = "auth.google_denied"
	MsgGoogleEmailUnverified MessageCode = "auth.google_email_unverified"
	MsgResetCodeInvalid      MessageCode = "auth.reset_code_invalid"
	MsgResetRateLimited      MessageCode = "auth.reset_rate_limited"
	MsgCaptchaFailed         MessageCode = "auth.captcha_failed"
	MsgAccountDisabled       MessageCode = "auth.account_disabled"
	MsgAccountDeleted        MessageCode = "auth.account_deleted"

	MsgUserNotFound   MessageCode = "user.not_found"
	MsgUserOwnsGroups MessageCode = "user.owns_groups"

	MsgVoiceRoomFull           MessageCode = "voice.room_full"
	MsgVoiceVideoLimitReached  MessageCode = "voice.video_limit_reached"
	MsgVoiceChannelNotVoice    MessageCode = "voice.channel_not_voice_capable"
	MsgVoiceQuotaExceeded      MessageCode = "voice.quota_exceeded"
	MsgVoiceRecordingAlready   MessageCode = "voice.recording_already_active"
	MsgVoiceRecordingNotActive MessageCode = "voice.recording_not_active"
	MsgVoiceWebhookInvalidSig  MessageCode = "voice.webhook_invalid_signature"
	MsgVoiceGuestNotAllowed    MessageCode = "voice.guest_not_allowed"
)

// ValidMessageCode reports whether code is an allowed i18n key shape from trusted services.
func ValidMessageCode(code MessageCode) bool {
	s := string(code)
	if s == "" {
		return false
	}

	return strings.HasPrefix(s, "error.") ||
		strings.HasPrefix(s, "auth.") ||
		strings.HasPrefix(s, "user.") ||
		strings.HasPrefix(s, "voice.")
}
