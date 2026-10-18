// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/google"
	"github.com/wichat/wichat/backend/ms-user/pkg/jwt"
	"github.com/wichat/wichat/backend/ms-user/pkg/password"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

const (
	accessTTL           = 15 * time.Minute
	refreshTTL          = 30 * 24 * time.Hour
	minPasswordLen      = 12
	maxPasswordLen      = 128
	maxDisplayNameRunes = 80
	maxUserAgentLen     = 512
	minJWTSecretLen     = 32
	googleProvider      = "google"
)

// Service is the auth use-case API (setup, password login, sessions, Google OAuth).
type Service interface {
	SetupStatus(ctx context.Context) (bool, error)
	Setup(ctx context.Context, in models.SetupInput) (*models.SessionResult, error)
	Login(ctx context.Context, in models.LoginInput) (*models.SessionResult, error)
	Refresh(ctx context.Context, refreshToken string, meta models.ClientMeta) (*models.SessionResult, error)
	Logout(ctx context.Context, refreshToken string) error
	Me(ctx context.Context, accessToken string) (*models.PublicUser, error)
	GoogleStart(ctx context.Context) (string, string, string, error)
	GoogleFinish(ctx context.Context, code, verifier string, meta models.ClientMeta) (*models.SessionResult, error)
}

type service struct {
	repo   repository.Repository
	log    logger.Logger
	secret []byte
	now    func() time.Time
	dummy  string
	google google.Config
}

// NewService builds the auth use cases. secret is the HMAC key and must be at least 32 bytes.
func NewService(repo repository.Repository, secret []byte, log logger.Logger, googleCfg google.Config) (Service, error) {
	if len(secret) < minJWTSecretLen {
		return nil, errors.New("auth: jwt secret must be at least 32 bytes")
	}

	dummy, err := password.DummyHash()
	if err != nil {
		return nil, err
	}

	copied := make([]byte, len(secret))
	copy(copied, secret)

	return service{
		repo:   repo,
		log:    log,
		secret: copied,
		now:    func() time.Time { return time.Now().UTC() },
		dummy:  dummy,
		google: google.Prepare(googleCfg),
	}, nil
}

// SetupStatus reports whether the instance still needs the first admin account.
func (s service) SetupStatus(ctx context.Context) (bool, error) {
	n, err := s.repo.CountUsers(ctx)
	if err != nil {
		return false, err
	}

	return n == 0, nil
}

// Setup creates the first user only; a second call returns setup closed.
func (s service) Setup(ctx context.Context, in models.SetupInput) (*models.SessionResult, error) {
	email, name, err := validateNewAccount(in.Email, in.Password, in.DisplayName)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteExpired(ctx, s.now()); err != nil {
		return nil, err
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return nil, exception.Internal(err)
	}

	now := s.now()
	id, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	user := models.User{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
		DisplayName:  name,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.CreateFirstUser(ctx, user); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, exception.Conflict(exception.MsgSetupClosed, nil)
		}

		return nil, err
	}

	return s.openSession(ctx, user, in.ClientMeta)
}

// Login checks email/password and opens a new session. Unknown emails still run argon2 against a dummy hash.
func (s service) Login(ctx context.Context, in models.LoginInput) (*models.SessionResult, error) {
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return nil, err
	}
	if !validPasswordLen(in.Password) {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "password"})
	}

	if err := s.repo.DeleteExpired(ctx, s.now()); err != nil {
		return nil, err
	}

	user, findErr := s.repo.FindUserByEmail(ctx, email)
	if findErr != nil && !errors.Is(findErr, repository.ErrNotFound) {
		return nil, findErr
	}

	hash := s.dummy
	if findErr == nil && user.PasswordHash != "" {
		hash = user.PasswordHash
	}
	if !password.Verify(hash, in.Password) || findErr != nil || user.PasswordHash == "" {
		return nil, exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
	}

	return s.openSession(ctx, user, in.ClientMeta)
}

// Refresh rotates the refresh token and issues a new access JWT for the same session id.
func (s service) Refresh(ctx context.Context, refreshToken string, _ models.ClientMeta) (*models.SessionResult, error) {
	if refreshToken == "" {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	if err := s.repo.DeleteExpired(ctx, s.now()); err != nil {
		return nil, err
	}

	rec, err := s.repo.FindRefreshByHash(ctx, token.HashRefresh(refreshToken))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}

	sess, err := s.repo.FindSession(ctx, rec.SessionID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}

	if sess.RevokedAt != nil || !sess.ExpiresAt.After(s.now()) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	// Presenting a refresh row that was already rotated is treated as token theft.
	if rec.UsedAt != nil {
		return s.reuse(ctx, sess.ID)
	}

	plain, nextHash, err := token.NewRefresh()
	if err != nil {
		return nil, exception.Internal(err)
	}
	nextID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	now := s.now()
	err = s.repo.MarkRefreshUsed(ctx, token.HashRefresh(refreshToken), nextHash, nextID, sess.ID, now.Add(refreshTTL), now)
	if errors.Is(err, repository.ErrAlreadyUsed) {
		return s.reuse(ctx, sess.ID)
	}
	if errors.Is(err, repository.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}

	user, err := s.repo.FindUserByID(ctx, sess.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}

	access, expiresIn, err := jwt.Issue(s.secret, user.ID, sess.ID, now, accessTTL)
	if err != nil {
		return nil, exception.Internal(err)
	}

	return &models.SessionResult{
		AccessToken:  access,
		ExpiresIn:    expiresIn,
		RefreshToken: plain,
		User:         user.Public(),
	}, nil
}

// Logout revokes the session tied to the refresh token, if any.
func (s service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	rec, err := s.repo.FindRefreshByHash(ctx, token.HashRefresh(refreshToken))
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	return s.repo.RevokeSession(ctx, rec.SessionID, s.now())
}

// Me loads the user for a valid access token and live session.
func (s service) Me(ctx context.Context, accessToken string) (*models.PublicUser, error) {
	if accessToken == "" {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	userID, sessionID, err := jwt.Parse(s.secret, accessToken)
	if err != nil {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	sess, err := s.repo.FindSession(ctx, sessionID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}

	if sess.UserID != userID || sess.RevokedAt != nil || !sess.ExpiresAt.After(s.now()) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}

	pub := user.Public()

	return &pub, nil
}

// GoogleStart returns the authorize URL, OAuth state, and PKCE verifier for the gateway cookie.
func (s service) GoogleStart(context.Context) (string, string, string, error) {
	if !google.Configured(s.google) {
		return "", "", "", exception.Unavailable(exception.MsgUnavailable, nil)
	}

	state, err := google.RandomToken()
	if err != nil {
		return "", "", "", exception.Internal(err)
	}
	verifier, challenge, err := google.PKCE()
	if err != nil {
		return "", "", "", exception.Internal(err)
	}

	return google.AuthorizeURL(s.google, state, challenge), state, verifier, nil
}

// GoogleFinish links or creates a user from Google and opens a session.
func (s service) GoogleFinish(ctx context.Context, code, verifier string, meta models.ClientMeta) (*models.SessionResult, error) {
	if !google.Configured(s.google) {
		return nil, exception.Unavailable(exception.MsgUnavailable, nil)
	}
	if code == "" || len(code) > 2048 || verifier == "" || len(verifier) > 128 {
		return nil, exception.Unauthenticated(exception.MsgGoogleDenied, nil)
	}

	profile, err := google.Exchange(ctx, s.google, code, verifier)
	if errors.Is(err, google.ErrDenied) {
		return nil, exception.Unauthenticated(exception.MsgGoogleDenied, nil)
	}
	if errors.Is(err, google.ErrUnavailable) {
		return nil, exception.Unavailable(exception.MsgUnavailable, nil)
	}
	if err != nil {
		return nil, exception.Internal(err)
	}
	if !profile.Verified {
		return nil, exception.Unauthenticated(exception.MsgGoogleEmailUnverified, nil)
	}
	email, err := normalizeEmail(profile.Email)
	if err != nil {
		return nil, exception.Unauthenticated(exception.MsgGoogleDenied, nil)
	}

	if err := s.repo.DeleteExpired(ctx, s.now()); err != nil {
		return nil, err
	}

	now := s.now()
	userID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}
	identityID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	user, err := s.repo.SaveGoogleUser(ctx, models.User{
		ID:           userID,
		Email:        email,
		PasswordHash: "",
		DisplayName:  google.DisplayName(profile.Name, email, maxDisplayNameRunes),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, identityID, googleProvider, profile.Subject)
	if err != nil {
		return nil, err
	}

	return s.openSession(ctx, user, meta)
}

// openSession persists a new session row and returns access + refresh tokens.
func (s service) openSession(ctx context.Context, user models.User, meta models.ClientMeta) (*models.SessionResult, error) {
	now := s.now()
	sessionID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	plain, hash, err := token.NewRefresh()
	if err != nil {
		return nil, exception.Internal(err)
	}
	refreshID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	session := models.Session{
		ID:         sessionID,
		UserID:     user.ID,
		ExpiresAt:  now.Add(refreshTTL),
		UserAgent:  clip(meta.UserAgent, maxUserAgentLen),
		ClientIP:   meta.ClientIP,
		LastUsedAt: now,
		CreatedAt:  now,
	}
	refresh := models.RefreshToken{
		ID:        refreshID,
		SessionID: sessionID,
		TokenHash: hash,
		CreatedAt: now,
	}
	if err := s.repo.InsertSession(ctx, session, refresh); err != nil {
		return nil, err
	}

	access, expiresIn, err := jwt.Issue(s.secret, user.ID, sessionID, now, accessTTL)
	if err != nil {
		return nil, exception.Internal(err)
	}

	return &models.SessionResult{
		AccessToken:  access,
		ExpiresIn:    expiresIn,
		RefreshToken: plain,
		User:         user.Public(),
	}, nil
}

// reuse revokes the session when a spent refresh token is presented again (rotation attack).
func (s service) reuse(ctx context.Context, sessionID string) (*models.SessionResult, error) {
	if s.log != nil {
		s.log.Warn("refresh token reuse", "session_id", sessionID)
	}
	if err := s.repo.RevokeSession(ctx, sessionID, s.now()); err != nil {
		return nil, err
	}

	return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
}

func validateNewAccount(email, password, displayName string) (string, string, error) {
	normalized, err := normalizeEmail(email)
	if err != nil {
		return "", "", err
	}
	if !validPasswordLen(password) {
		return "", "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "password"})
	}

	name := strings.TrimSpace(displayName)
	if name == "" || utf8.RuneCountInString(name) > maxDisplayNameRunes {
		return "", "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "display_name"})
	}

	return normalized, name, nil
}

func normalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if !validEmail(normalized) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "email"})
	}

	return normalized, nil
}

func validPasswordLen(password string) bool {
	return len(password) >= minPasswordLen && len(password) <= maxPasswordLen
}

func validEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 || strings.ContainsAny(email, " \t\r\n") {
		return false
	}

	at := strings.IndexByte(email, '@')
	if at <= 0 || at != strings.LastIndexByte(email, '@') || at >= len(email)-1 {
		return false
	}

	return strings.Contains(email[at+1:], ".")
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return s[:n]
}
