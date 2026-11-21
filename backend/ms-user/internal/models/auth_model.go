package models

import (
	"errors"
	"time"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

// User is a stored account, including the password hash.
type User struct {
	ID                  string
	Email               string
	Username            string
	PasswordHash        string
	DisplayName         string
	AvatarURL           string
	Pronouns            string
	Bio                 string
	BannerColor         string
	StatusNoteText      string
	StatusNoteEmoji     string
	StatusNoteExpiresAt *time.Time
	Locale              string
	TimeFormat          string
	BirthDate           *time.Time
	IsGuest             bool
	IsInstanceOwner     bool
	DisabledAt          *time.Time
	DeletedAt           *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// SetupStatus is the public instance policy snapshot for the auth UI.
type SetupStatus struct {
	RegistrationMode string
}

// CompleteInstanceSetupInput is the owner setup wizard payload (first group + instance policy).
type CompleteInstanceSetupInput struct {
	Name             string
	LogoURL          string
	RegistrationMode string
}

// InstanceSetupPayload is the validated wizard data passed to the repository.
type InstanceSetupPayload struct {
	GroupID          string
	GroupPublicID    string
	GroupName        string
	GroupLogoURL     string
	RegistrationMode string
}

// CheckUsernameResult is the availability check outcome for registration.
type CheckUsernameResult struct {
	Available  bool
	Normalized string
	Reason     string
}

// Public returns the account fields safe to send to clients.
func (u User) Public() PublicUser {
	noteText := u.StatusNoteText
	noteEmoji := u.StatusNoteEmoji
	noteExpiresAt := u.StatusNoteExpiresAt
	if noteExpiresAt != nil && !noteExpiresAt.After(time.Now().UTC()) {
		noteText = ""
		noteEmoji = ""
		noteExpiresAt = nil
	}

	timeFormat := u.TimeFormat
	if timeFormat == "" {
		timeFormat = "auto"
	}

	return PublicUser{
		ID:                  u.ID,
		Email:               u.Email,
		Username:            u.Username,
		DisplayName:         u.DisplayName,
		AvatarURL:           u.AvatarURL,
		Pronouns:            u.Pronouns,
		Bio:                 u.Bio,
		BannerColor:         u.BannerColor,
		StatusNoteText:      noteText,
		StatusNoteEmoji:     noteEmoji,
		StatusNoteExpiresAt: noteExpiresAt,
		Locale:              u.Locale,
		TimeFormat:          timeFormat,
		CreatedAt:           u.CreatedAt,
		HasPassword:         u.PasswordHash != "",
	}
}

// PublicUser is the account view returned to clients.
type PublicUser struct {
	ID                  string
	Email               string
	Username            string
	DisplayName         string
	AvatarURL           string
	Pronouns            string
	Bio                 string
	BannerColor         string
	StatusNoteText      string
	StatusNoteEmoji     string
	StatusNoteExpiresAt *time.Time
	Locale              string
	TimeFormat          string
	CreatedAt           time.Time
	HasPassword         bool
}

// UserStatusNotePatch updates or clears the user's status note on users.
type UserStatusNotePatch struct {
	Text      string
	Emoji     string
	ExpiresAt *time.Time
}

// UserPreferencesPatch updates the user's display locale and time format on users.
type UserPreferencesPatch struct {
	Locale     string
	TimeFormat string
}

// UserProfilePatch updates user-editable profile fields on users.
type UserProfilePatch struct {
	DisplayName string
	Pronouns    string
	Bio         string
	BannerColor string
	ClearAvatar bool
}

// ClientMeta is the device hint stored on a session.
type ClientMeta struct {
	UserAgent     string
	ClientIP      string
	DeviceLabel   string
	Platform      string
	LocationLabel string
}

// SetupInput is the legacy first-account form (alias of RegisterInput fields).
type SetupInput struct {
	Email       string
	Password    string
	DisplayName string
	ClientMeta
}

// RegisterInput is the member registration form.
type RegisterInput struct {
	Email       string
	Username    string
	Password    string
	DisplayName string
	BirthDate   string
	Locale      string
	InviteToken string
	ClientMeta
}

// LoginInput is a login identifier and password.
type LoginInput struct {
	Login    string
	Password string
	Locale   string
	ClientMeta
}

// Session is one device login. Its id stays stable across refresh rotation.
type Session struct {
	ID               string
	UserID           string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	UserAgent        string
	ClientIP         string
	DeviceLabel      string
	Platform         string
	LocationLabel    string
	RefreshID        string
	RefreshTokenHash []byte
	LastUsedAt       time.Time
	CreatedAt        time.Time
}

// RefreshToken is a stored refresh secret. Only the hash is persisted.
type RefreshToken struct {
	ID        string
	SessionID string
	TokenHash []byte
	UsedAt    *time.Time
	CreatedAt time.Time
}

// SessionResult is a new access token plus the plaintext refresh token.
type SessionResult struct {
	AccessToken  string
	ExpiresIn    int32
	RefreshToken string
	User         PublicUser
}

// AuthChallenge is the email OTP or MFA step before a session is issued.
type AuthChallenge struct {
	ChallengeID           string
	EmailHint             string
	ResendCooldownSeconds int32
	MFARequired           bool
	MFAMethods            []string
	WebAuthnOptionsJSON   string
}

// PasswordAuthResult is either a session or an email OTP challenge.
type PasswordAuthResult struct {
	Session   *SessionResult
	Challenge *AuthChallenge
}

// RegisterPending holds signup data in Redis until email OTP is verified.
type RegisterPending struct {
	Email        string `json:"email"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	DisplayName  string `json:"display_name"`
	BirthDate    string `json:"birth_date"`
	Locale       string `json:"locale"`
	CodeHMAC     string `json:"code_hmac"`
	Attempts     int    `json:"attempts"`
	OtpVerified  bool   `json:"otp_verified,omitempty"`
	UserAgent    string `json:"user_agent"`
	ClientIP     string `json:"client_ip"`
}

// PasswordResetOTP is the Redis payload for a password-reset email code.
type PasswordResetOTP struct {
	UserID   string `json:"user_id"`
	CodeHMAC string `json:"code_hmac"`
	Attempts int    `json:"attempts"`
}

// PasswordResetTokenRecord is the Redis payload for a one-time password reset token.
type PasswordResetTokenRecord struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// ResetRateLimitError carries how long to wait before another auth email code send.
type ResetRateLimitError struct {
	RetryAfter time.Duration
}

func (e *ResetRateLimitError) Error() string {
	return "auth: reset rate limited"
}

func (e *ResetRateLimitError) Unwrap() error {
	return constants.ErrResetRateLimited
}

// ResetRateLimitRetryAfter returns remaining wait when err is a resend cooldown.
func ResetRateLimitRetryAfter(err error) (time.Duration, bool) {
	var rl *ResetRateLimitError
	if errors.As(err, &rl) {
		return rl.RetryAfter, true
	}
	if errors.Is(err, constants.ErrResetRateLimited) {
		return 0, true
	}

	return 0, false
}

// LoginPending holds a verified-password login in Redis awaiting email OTP.
type LoginPending struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Locale      string `json:"locale"`
	CodeHMAC    string `json:"code_hmac"`
	Attempts    int    `json:"attempts"`
	UserAgent   string `json:"user_agent"`
	ClientIP    string `json:"client_ip"`
}

// EmailChangePending holds an email-change OTP challenge in Redis.
type EmailChangePending struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	CodeHMAC string `json:"code_hmac"`
	Attempts int    `json:"attempts"`
}

// EmailChangeTokenRecord holds a verified one-time email-change token in Redis.
type EmailChangeTokenRecord struct {
	UserID string `json:"user_id"`
}

// LoginMFAPending holds a password-verified login awaiting MFA (WebAuthn, TOTP, or recovery code).
type LoginMFAPending struct {
	UserID              string   `json:"user_id"`
	Email               string   `json:"email"`
	Methods             []string `json:"methods"`
	Attempts            int      `json:"attempts"`
	WebAuthnSessionJSON string   `json:"webauthn_session_json,omitempty"`
	UserAgent           string   `json:"user_agent"`
	ClientIP            string   `json:"client_ip"`
}

// TOTPSetupPending holds an unconfirmed TOTP secret in Redis.
type TOTPSetupPending struct {
	UserID string `json:"user_id"`
	Secret string `json:"secret"`
}

// WebAuthnRegisterPending holds an ephemeral WebAuthn registration ceremony in Redis.
type WebAuthnRegisterPending struct {
	UserID      string `json:"user_id"`
	KeyName     string `json:"key_name"`
	SessionJSON string `json:"session_json"`
}

// MFATOTP represents an enabled TOTP configuration in PostgreSQL.
type MFATOTP struct {
	UserID          string
	SecretEncrypted []byte
	LastUsedStep    int64
	EnabledAt       time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// WebAuthnCredential represents a registered FIDO2/WebAuthn security key in PostgreSQL.
type WebAuthnCredential struct {
	ID              string
	UserID          string
	Name            string
	CredentialID    []byte
	PublicKey       []byte
	AttestationType string
	AAGUID          []byte
	SignCount       uint32
	CloneWarning    bool
	Transports      []string
	BackupEligible  bool
	BackupState     bool
	LastUsedAt      *time.Time
	CreatedAt       time.Time
}

// MFAStatus is the user's enrolled MFA overview.
type MFAStatus struct {
	TOTPEnabled            bool
	SecurityKeys           []WebAuthnCredential
	RecoveryCodesRemaining int32
}

// TOTPSetupStart is returned when initiating TOTP enrollment.
type TOTPSetupStart struct {
	SetupToken string
	Secret     string
	OtpauthURI string
}

// WebAuthnRegisterBegin is returned when initiating Security Key enrollment.
type WebAuthnRegisterBegin struct {
	ChallengeID string
	OptionsJSON string
}

// UserRelationship represents a user's ignore or block entry against another user.
type UserRelationship struct {
	UserID       string
	TargetUserID string
	Kind         string
	CreatedAt    time.Time
}
