package models

import (
	"errors"
	"time"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

// User is a stored account, including the password hash.
type User struct {
	ID              string
	Email           string
	Username        string
	PasswordHash    string
	DisplayName     string
	BirthDate       *time.Time
	IsGuest         bool
	IsInstanceOwner bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// InstanceSettings is the singleton instance policy row.
type InstanceSettings struct {
	UsernameEnabled  bool
	RegistrationMode string
	GuestEnabled     bool
	SetupCompleted   bool
}

// Group is a workspace-equivalent chat unit (product name: Group).
type Group struct {
	ID        string
	Name      string
	LogoURL   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PublicGroup is the group view returned to clients.
type PublicGroup struct {
	ID      string
	Name    string
	LogoURL string
}

// Public returns client-safe group fields.
func (g Group) Public() PublicGroup {
	return PublicGroup{ID: g.ID, Name: g.Name, LogoURL: g.LogoURL}
}

// GroupMember links a user to a group with a preset role.
type GroupMember struct {
	GroupID   string
	UserID    string
	Role      string
	CreatedAt time.Time
}

// SetupStatus is the public instance policy snapshot for the auth UI.
type SetupStatus struct {
	RegistrationMode string
	GuestEnabled     bool
	UsernameEnabled  bool
}

// CompleteInstanceSetupInput is the owner setup wizard payload (first group + instance policy).
type CompleteInstanceSetupInput struct {
	Name             string
	LogoURL          string
	UsernameEnabled  bool
	RegistrationMode string
	GuestEnabled     bool
}

// InstanceSetupPayload is the validated wizard data passed to the repository.
type InstanceSetupPayload struct {
	GroupID          string
	GroupName        string
	GroupLogoURL     string
	UsernameEnabled  bool
	RegistrationMode string
	GuestEnabled     bool
}

// Public returns the account fields safe to send to clients.
func (u User) Public() PublicUser {
	return PublicUser{
		ID:          u.ID,
		Email:       u.Email,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		IsGuest:     u.IsGuest,
	}
}

// PublicUser is the account view returned to clients.
type PublicUser struct {
	ID          string
	Email       string
	Username    string
	DisplayName string
	IsGuest     bool
}

// ClientMeta is the device hint stored on a session.
type ClientMeta struct {
	UserAgent string
	ClientIP  string
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
	Password    string
	DisplayName string
	BirthDate   string
	Locale      string
	ClientMeta
}

// GuestRegisterInput is the guest registration form.
type GuestRegisterInput struct {
	DisplayName string
	BirthDate   string
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
	ID         string
	UserID     string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	UserAgent  string
	ClientIP   string
	LastUsedAt time.Time
	CreatedAt  time.Time
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

// AuthChallenge is the email OTP step before a session is issued.
type AuthChallenge struct {
	ChallengeID           string
	EmailHint             string
	ResendCooldownSeconds int32
}

// PasswordAuthResult is either a session or an email OTP challenge.
type PasswordAuthResult struct {
	Session   *SessionResult
	Challenge *AuthChallenge
}

// RegisterPending holds signup data in Redis until email OTP is verified.
type RegisterPending struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	DisplayName  string `json:"display_name"`
	BirthDate    string `json:"birth_date"`
	Locale       string `json:"locale"`
	CodeHMAC     string `json:"code_hmac"`
	Attempts     int    `json:"attempts"`
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
