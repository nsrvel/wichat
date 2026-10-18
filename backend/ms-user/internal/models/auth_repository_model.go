// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package models

import (
	"time"

	"github.com/uptrace/bun"
)

// UserRow maps auth.users for Bun.
type UserRow struct {
	bun.BaseModel `bun:"table:auth.users"`

	ID           string    `bun:"id,pk"`
	Email        string    `bun:"email"`
	PasswordHash string    `bun:"password_hash"`
	DisplayName  string    `bun:"display_name"`
	CreatedAt    time.Time `bun:"created_at"`
	UpdatedAt    time.Time `bun:"updated_at"`
}

func UserRowFrom(u User) UserRow {
	return UserRow{
		ID: u.ID, Email: u.Email, PasswordHash: u.PasswordHash,
		DisplayName: u.DisplayName, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}

func (r UserRow) ToModel() User {
	return User{
		ID: r.ID, Email: r.Email, PasswordHash: r.PasswordHash,
		DisplayName: r.DisplayName, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
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
