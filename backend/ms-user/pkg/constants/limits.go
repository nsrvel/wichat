package constants

import "time"

// Account and credential limits.
const (
	MinPasswordLen      = 12
	MaxPasswordLen      = 128
	MinJWTSecretLen     = 32
	MaxDisplayNameRunes = 80
	MaxUserAgentLen     = 512
	MinAgeYears         = 13
)

// Group and instance setup limits.
const (
	MaxGroupNameRunes        = 80
	MaxGroupDescriptionRunes = 2000
	MaxLogoURLRunes          = 2048
	MaxGroupIconBytes        = 5 * 1024 * 1024
	MaxGroupBannerBytes      = 8 * 1024 * 1024
)

// AllowedGroupBannerColors are solid banner palette ids (matches web presets minus image).
var AllowedGroupBannerColors = map[string]struct{}{
	"default": {},
	"blossom": {},
	"volcano": {},
	"sunset":  {},
	"gold":    {},
	"grape":   {},
	"sky":     {},
	"forest":  {},
	"onyx":    {},
}

// Group invite and member list limits.
const (
	DefaultMemberListLimit = 20
	MaxMemberListLimit     = 50
	DefaultUserSearchLimit = 20
	MaxUserSearchLimit     = 20
	MaxInviteBatchSize     = 25
	GroupInviteTTL         = 7 * 24 * time.Hour
	GroupShareLinkTTL      = 30 * 24 * time.Hour
)

// Session token lifetimes.
const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 30 * 24 * time.Hour
	// RefreshReuseRevokeGrace: parallel tab refresh or a failed response after Rotate must not revoke the session.
	RefreshReuseRevokeGrace = 30 * time.Second
)

// Email OTP and password-reset Redis TTLs and attempt caps.
const (
	EmailAuthOTPTTL          = 10 * time.Minute
	EmailAuthMaxAttempts     = 5
	PasswordResetOTPTTL      = 10 * time.Minute
	PasswordResetTokenTTL    = 15 * time.Minute
	PasswordResetMaxAttempts = 5
	OTPLength                = 6
)

// GoogleOAuthProvider is the identity provider key stored for Google link rows.
const GoogleOAuthProvider = "google"
