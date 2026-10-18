// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package models

import "time"

// User is a stored account, including the password hash.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Public returns the account fields safe to send to clients.
func (u User) Public() PublicUser {
	return PublicUser{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName}
}

// PublicUser is the account view returned to clients.
type PublicUser struct {
	ID          string
	Email       string
	DisplayName string
}

// ClientMeta is the device hint stored on a session.
type ClientMeta struct {
	UserAgent string
	ClientIP  string
}

// SetupInput is the first-account form.
type SetupInput struct {
	Email       string
	Password    string
	DisplayName string
	ClientMeta
}

// LoginInput is an email and password login.
type LoginInput struct {
	Email    string
	Password string
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
