package models

import (
	"time"

	"github.com/uptrace/bun"
)

// UserRow maps auth.users for Bun.
type UserRow struct {
	bun.BaseModel `bun:"table:auth.users"`

	ID              string     `bun:"id,pk"`
	Email           *string    `bun:"email"`
	Username        *string    `bun:"username"`
	PasswordHash    *string    `bun:"password_hash"`
	DisplayName     string     `bun:"display_name"`
	BirthDate       *time.Time `bun:"birth_date,nullzero"`
	IsGuest         bool       `bun:"is_guest"`
	IsInstanceOwner bool       `bun:"is_instance_owner"`
	CreatedAt       time.Time  `bun:"created_at"`
	UpdatedAt       time.Time  `bun:"updated_at"`
}

// InstanceSettingsRow maps auth.instance_settings for Bun.
type InstanceSettingsRow struct {
	bun.BaseModel `bun:"table:auth.instance_settings"`

	ID               int16     `bun:"id,pk"`
	UsernameEnabled  bool      `bun:"username_enabled"`
	RegistrationMode string    `bun:"registration_mode"`
	GuestEnabled     bool      `bun:"guest_enabled"`
	SetupCompleted   bool      `bun:"setup_completed"`
	CreatedAt        time.Time `bun:"created_at"`
	UpdatedAt        time.Time `bun:"updated_at"`
}

func (r InstanceSettingsRow) ToModel() InstanceSettings {
	return InstanceSettings{
		UsernameEnabled:  r.UsernameEnabled,
		RegistrationMode: r.RegistrationMode,
		GuestEnabled:     r.GuestEnabled,
		SetupCompleted:   r.SetupCompleted,
	}
}

// GroupRow maps auth.groups for Bun.
type GroupRow struct {
	bun.BaseModel `bun:"table:auth.groups"`

	ID        string    `bun:"id,pk"`
	Name      string    `bun:"name"`
	LogoURL   string    `bun:"logo_url"`
	CreatedAt time.Time `bun:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"`
}

func GroupRowFrom(g Group) GroupRow {
	return GroupRow{
		ID: g.ID, Name: g.Name, LogoURL: g.LogoURL, CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
	}
}

func (r GroupRow) ToModel() Group {
	return Group{
		ID: r.ID, Name: r.Name, LogoURL: r.LogoURL, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

// GroupMemberRow maps auth.group_members for Bun.
type GroupMemberRow struct {
	bun.BaseModel `bun:"table:auth.group_members"`

	GroupID   string    `bun:"group_id,pk"`
	UserID    string    `bun:"user_id,pk"`
	Role      string    `bun:"role"`
	CreatedAt time.Time `bun:"created_at"`
}

func GroupMemberRowFrom(m GroupMember) GroupMemberRow {
	return GroupMemberRow{
		GroupID: m.GroupID, UserID: m.UserID, Role: m.Role, CreatedAt: m.CreatedAt,
	}
}

func UserRowFrom(u User) UserRow {
	var email, username, pass *string
	if u.Email != "" {
		email = &u.Email
	}
	if u.Username != "" {
		username = &u.Username
	}
	if u.PasswordHash != "" {
		pass = &u.PasswordHash
	}

	return UserRow{
		ID: u.ID, Email: email, Username: username, PasswordHash: pass,
		DisplayName: u.DisplayName, BirthDate: u.BirthDate, IsGuest: u.IsGuest,
		IsInstanceOwner: u.IsInstanceOwner,
		CreatedAt:       u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}

func (r UserRow) ToModel() User {
	u := User{
		DisplayName:     r.DisplayName,
		BirthDate:       r.BirthDate,
		IsGuest:         r.IsGuest,
		IsInstanceOwner: r.IsInstanceOwner,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		ID:              r.ID,
	}
	if r.Email != nil {
		u.Email = *r.Email
	}
	if r.Username != nil {
		u.Username = *r.Username
	}
	if r.PasswordHash != nil {
		u.PasswordHash = *r.PasswordHash
	}

	return u
}

// SessionRow maps auth.sessions for Bun.
type SessionRow struct {
	bun.BaseModel `bun:"table:auth.sessions"`

	ID         string     `bun:"id,pk"`
	UserID     string     `bun:"user_id"`
	ExpiresAt  time.Time  `bun:"expires_at"`
	RevokedAt  *time.Time `bun:"revoked_at"`
	UserAgent  string     `bun:"user_agent"`
	ClientIP   string     `bun:"client_ip"`
	LastUsedAt time.Time  `bun:"last_used_at"`
	CreatedAt  time.Time  `bun:"created_at"`
}

func SessionRowFrom(s Session) SessionRow {
	return SessionRow{
		ID: s.ID, UserID: s.UserID, ExpiresAt: s.ExpiresAt, RevokedAt: s.RevokedAt,
		UserAgent: s.UserAgent, ClientIP: s.ClientIP, LastUsedAt: s.LastUsedAt, CreatedAt: s.CreatedAt,
	}
}

func (r SessionRow) ToModel() Session {
	return Session{
		ID: r.ID, UserID: r.UserID, ExpiresAt: r.ExpiresAt, RevokedAt: r.RevokedAt,
		UserAgent: r.UserAgent, ClientIP: r.ClientIP, LastUsedAt: r.LastUsedAt, CreatedAt: r.CreatedAt,
	}
}

// RefreshRow maps auth.refresh_tokens for Bun.
type RefreshRow struct {
	bun.BaseModel `bun:"table:auth.refresh_tokens"`

	ID        string     `bun:"id,pk"`
	SessionID string     `bun:"session_id"`
	TokenHash []byte     `bun:"token_hash"`
	UsedAt    *time.Time `bun:"used_at"`
	CreatedAt time.Time  `bun:"created_at"`
}

func RefreshRowFrom(t RefreshToken) RefreshRow {
	return RefreshRow{
		ID: t.ID, SessionID: t.SessionID, TokenHash: t.TokenHash,
		UsedAt: t.UsedAt, CreatedAt: t.CreatedAt,
	}
}

func (r RefreshRow) ToModel() RefreshToken {
	return RefreshToken{
		ID: r.ID, SessionID: r.SessionID, TokenHash: r.TokenHash,
		UsedAt: r.UsedAt, CreatedAt: r.CreatedAt,
	}
}

// IdentityRow maps auth.identities for Bun.
type IdentityRow struct {
	bun.BaseModel `bun:"table:auth.identities"`

	ID        string    `bun:"id,pk"`
	UserID    string    `bun:"user_id"`
	Provider  string    `bun:"provider"`
	Subject   string    `bun:"subject"`
	CreatedAt time.Time `bun:"created_at"`
}
