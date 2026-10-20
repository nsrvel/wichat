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
	MaxGroupNameRunes = 80
	MaxLogoURLRunes   = 2048
)

// Session token lifetimes.
const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 30 * 24 * time.Hour
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
