package models

import (
	"time"

	"github.com/uptrace/bun"
)

// UserRow maps public.users for Bun.
type UserRow struct {
	bun.BaseModel `bun:"table:users"`

	ID              string     `bun:"id,pk,type:uuid"`
	Email           *string    `bun:"email"`
	Username        string     `bun:"username"`
	DisplayName     string     `bun:"display_name"`
	AvatarURL       string     `bun:"avatar_url"`
	Pronouns        string     `bun:"pronouns"`
	Bio             string     `bun:"bio"`
	BannerColor     string     `bun:"banner_color"`
	BirthDate       *time.Time `bun:"birth_date,nullzero"`
	IsGuest         bool       `bun:"is_guest"`
	IsInstanceOwner bool       `bun:"is_instance_owner"`
	DisabledAt      *time.Time `bun:"disabled_at,nullzero"`
	DeletedAt       *time.Time `bun:"deleted_at,nullzero"`
	CreatedAt       time.Time  `bun:"created_at"`
	UpdatedAt       time.Time  `bun:"updated_at"`
}

// UserCredentialRow maps user_credentials for Bun.
type UserCredentialRow struct {
	bun.BaseModel `bun:"table:user_credentials"`

	UserID       string    `bun:"user_id,pk,type:uuid"`
	PasswordHash string    `bun:"password_hash"`
	UpdatedAt    time.Time `bun:"updated_at"`
}

// UserConnectionRow maps user_connections for Bun.
type UserConnectionRow struct {
	bun.BaseModel `bun:"table:user_connections"`

	ID        string    `bun:"id,pk,type:uuid"`
	UserID    string    `bun:"user_id,type:uuid"`
	Kind      string    `bun:"kind"`
	Provider  string    `bun:"provider"`
	Subject   string    `bun:"subject"`
	Scopes    []string  `bun:"scopes,array"`
	CreatedAt time.Time `bun:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"`
}

// InstanceSettingsRow maps instance_settings for Bun.
type InstanceSettingsRow struct {
	bun.BaseModel `bun:"table:instance_settings"`

	ID               int16     `bun:"id,pk"`
	RegistrationMode string    `bun:"registration_mode"`
	SetupCompleted   bool      `bun:"setup_completed"`
	CreatedAt        time.Time `bun:"created_at"`
	UpdatedAt        time.Time `bun:"updated_at"`
}

func (r InstanceSettingsRow) ToModel() InstanceSettings {
	return InstanceSettings{
		RegistrationMode: r.RegistrationMode,
		SetupCompleted:   r.SetupCompleted,
	}
}

func UserRowFrom(u User) UserRow {
	var email *string
	if u.Email != "" {
		email = &u.Email
	}

	return UserRow{
		ID: u.ID, Email: email, Username: u.Username,
		DisplayName: u.DisplayName, AvatarURL: u.AvatarURL,
		Pronouns: u.Pronouns, Bio: u.Bio, BannerColor: u.BannerColor,
		BirthDate: u.BirthDate, IsGuest: u.IsGuest,
		IsInstanceOwner: u.IsInstanceOwner,
		DisabledAt:      u.DisabledAt, DeletedAt: u.DeletedAt,
		CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}

func (r UserRow) ToModel(passwordHash string) User {
	u := User{
		Username:        r.Username,
		DisplayName:     r.DisplayName,
		AvatarURL:       r.AvatarURL,
		Pronouns:        r.Pronouns,
		Bio:             r.Bio,
		BannerColor:     r.BannerColor,
		BirthDate:       r.BirthDate,
		IsGuest:         r.IsGuest,
		IsInstanceOwner: r.IsInstanceOwner,
		DisabledAt:      r.DisabledAt,
		DeletedAt:       r.DeletedAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		ID:              r.ID,
		PasswordHash:    passwordHash,
	}
	if r.Email != nil {
		u.Email = *r.Email
	}

	return u
}

// SessionRow maps user_sessions for Bun.
type SessionRow struct {
	bun.BaseModel `bun:"table:user_sessions"`

	ID               string     `bun:"id,pk,type:uuid"`
	UserID           string     `bun:"user_id,type:uuid"`
	ExpiresAt        time.Time  `bun:"expires_at"`
	RevokedAt        *time.Time `bun:"revoked_at"`
	UserAgent        string     `bun:"user_agent"`
	ClientIP         string     `bun:"client_ip"`
	DeviceLabel      string     `bun:"device_label"`
	Platform         string     `bun:"platform"`
	LocationLabel    string     `bun:"location_label"`
	RefreshID        string     `bun:"refresh_id"`
	RefreshTokenHash []byte     `bun:"refresh_token_hash"`
	LastUsedAt       time.Time  `bun:"last_used_at"`
	CreatedAt        time.Time  `bun:"created_at"`
}

func SessionRowFrom(s Session) SessionRow {
	return SessionRow{
		ID: s.ID, UserID: s.UserID, ExpiresAt: s.ExpiresAt, RevokedAt: s.RevokedAt,
		UserAgent: s.UserAgent, ClientIP: s.ClientIP,
		DeviceLabel: s.DeviceLabel, Platform: s.Platform, LocationLabel: s.LocationLabel,
		RefreshID: s.RefreshID, RefreshTokenHash: s.RefreshTokenHash,
		LastUsedAt: s.LastUsedAt, CreatedAt: s.CreatedAt,
	}
}

func (r SessionRow) ToModel() Session {
	return Session{
		ID: r.ID, UserID: r.UserID, ExpiresAt: r.ExpiresAt, RevokedAt: r.RevokedAt,
		UserAgent: r.UserAgent, ClientIP: r.ClientIP,
		DeviceLabel: r.DeviceLabel, Platform: r.Platform, LocationLabel: r.LocationLabel,
		RefreshID: r.RefreshID, RefreshTokenHash: r.RefreshTokenHash,
		LastUsedAt: r.LastUsedAt, CreatedAt: r.CreatedAt,
	}
}

// MFATOTPRow maps user_mfa_totp for Bun.
type MFATOTPRow struct {
	bun.BaseModel `bun:"table:user_mfa_totp"`

	UserID          string    `bun:"user_id,pk,type:uuid"`
	SecretEncrypted []byte    `bun:"secret_encrypted"`
	LastUsedStep    int64     `bun:"last_used_step"`
	EnabledAt       time.Time `bun:"enabled_at"`
	CreatedAt       time.Time `bun:"created_at"`
	UpdatedAt       time.Time `bun:"updated_at"`
}

func (r MFATOTPRow) ToModel() MFATOTP {
	return MFATOTP{
		UserID:          r.UserID,
		SecretEncrypted: r.SecretEncrypted,
		LastUsedStep:    r.LastUsedStep,
		EnabledAt:       r.EnabledAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

// WebAuthnCredentialRow maps user_webauthn_credentials for Bun.
type WebAuthnCredentialRow struct {
	bun.BaseModel `bun:"table:user_webauthn_credentials"`

	ID              string     `bun:"id,pk,type:uuid"`
	UserID          string     `bun:"user_id,type:uuid"`
	Name            string     `bun:"name"`
	CredentialID    []byte     `bun:"credential_id"`
	PublicKey       []byte     `bun:"public_key"`
	AttestationType string     `bun:"attestation_type"`
	AAGUID          []byte     `bun:"aaguid"`
	SignCount       int64      `bun:"sign_count"`
	CloneWarning    bool       `bun:"clone_warning"`
	Transports      []string   `bun:"transports,array"`
	BackupEligible  bool       `bun:"backup_eligible"`
	BackupState     bool       `bun:"backup_state"`
	LastUsedAt      *time.Time `bun:"last_used_at"`
	CreatedAt       time.Time  `bun:"created_at"`
}

func (r WebAuthnCredentialRow) ToModel() WebAuthnCredential {
	return WebAuthnCredential{
		ID:              r.ID,
		UserID:          r.UserID,
		Name:            r.Name,
		CredentialID:    r.CredentialID,
		PublicKey:       r.PublicKey,
		AttestationType: r.AttestationType,
		AAGUID:          r.AAGUID,
		SignCount:       uint32(max(0, r.SignCount)), //nolint:gosec // G115: sign_count is non-negative uint32 in WebAuthn
		CloneWarning:    r.CloneWarning,
		Transports:      r.Transports,
		BackupEligible:  r.BackupEligible,
		BackupState:     r.BackupState,
		LastUsedAt:      r.LastUsedAt,
		CreatedAt:       r.CreatedAt,
	}
}

// MFARecoveryCodeRow maps user_mfa_recovery_codes for Bun.
type MFARecoveryCodeRow struct {
	bun.BaseModel `bun:"table:user_mfa_recovery_codes"`

	ID        string     `bun:"id,pk,type:uuid"`
	UserID    string     `bun:"user_id,type:uuid"`
	CodeHash  []byte     `bun:"code_hash"`
	UsedAt    *time.Time `bun:"used_at"`
	CreatedAt time.Time  `bun:"created_at"`
}
