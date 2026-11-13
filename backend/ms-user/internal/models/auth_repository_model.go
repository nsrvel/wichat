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
	BirthDate       *time.Time `bun:"birth_date,nullzero"`
	IsGuest         bool       `bun:"is_guest"`
	IsInstanceOwner bool       `bun:"is_instance_owner"`
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
		DisplayName: u.DisplayName, AvatarURL: u.AvatarURL, BirthDate: u.BirthDate, IsGuest: u.IsGuest,
		IsInstanceOwner: u.IsInstanceOwner,
		CreatedAt:       u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}

func (r UserRow) ToModel(passwordHash string) User {
	u := User{
		Username:        r.Username,
		DisplayName:     r.DisplayName,
		AvatarURL:       r.AvatarURL,
		BirthDate:       r.BirthDate,
		IsGuest:         r.IsGuest,
		IsInstanceOwner: r.IsInstanceOwner,
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
