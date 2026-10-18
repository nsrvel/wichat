// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

// Package google implements OAuth2 authorization URL, PKCE, and userinfo fetch for Google sign-in.
package google

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	// ErrDenied is returned when Google rejects the code or userinfo is invalid.
	ErrDenied = errors.New("google: denied")
	// ErrUnavailable is returned when the HTTP client cannot reach Google.
	ErrUnavailable = errors.New("google: unavailable")
)

// Config is the OAuth app credentials and HTTP transport. Empty ClientID disables Google sign-in.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthorizeURL string
	TokenURL     string
	UserInfoURL  string
	HTTP         *http.Client
}

// Prepare trims credentials and sets a default HTTP client when nil.
func Prepare(c Config) Config {
	c.ClientID = strings.TrimSpace(c.ClientID)
	c.ClientSecret = strings.TrimSpace(c.ClientSecret)
	c.RedirectURL = strings.TrimSpace(c.RedirectURL)
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 10 * time.Second}
	}

	return c
}

// Profile is the verified identity returned after code exchange.
type Profile struct {
	Subject  string
	Email    string
	Name     string
	Verified bool
}

// Configured reports whether Google sign-in can run.
func Configured(c Config) bool {
	return c.ClientID != "" && c.ClientSecret != "" && c.RedirectURL != ""
}

// RandomToken returns a URL-safe opaque string for OAuth state.
func RandomToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

// PKCE returns a verifier and S256 code challenge for the authorize request.
func PKCE() (verifier, challenge string, err error) {
	verifier, err = RandomToken()
	if err != nil {
		return "", "", err
	}
	sum := sha256.Sum256([]byte(verifier))

	return verifier, base64.RawURLEncoding.EncodeToString(sum[:]), nil
}

// AuthorizeURL builds the Google consent URL with PKCE and state.
func AuthorizeURL(c Config, state, challenge string) string {
	authorize := c.AuthorizeURL
	if authorize == "" {
		authorize = "https://accounts.google.com/o/oauth2/v2/auth"
	}
	query := url.Values{}
	query.Set("client_id", c.ClientID)
	query.Set("redirect_uri", c.RedirectURL)
	query.Set("response_type", "code")
	query.Set("scope", "openid email profile")
	query.Set("state", state)
	query.Set("code_challenge", challenge)
	query.Set("code_challenge_method", "S256")
	query.Set("prompt", "select_account")

	return authorize + "?" + query.Encode()
}

// Exchange trades an authorization code for a verified profile.
func Exchange(ctx context.Context, c Config, code, verifier string) (Profile, error) {
	tokenURL := c.TokenURL
	if tokenURL == "" {
		tokenURL = "https://oauth2.googleapis.com/token"
	}
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("redirect_uri", c.RedirectURL)
	form.Set("grant_type", "authorization_code")
	form.Set("code_verifier", verifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Profile{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenBody, err := do(c.HTTP, req)
	if err != nil {
		return Profile{}, err
	}

	var token struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(tokenBody, &token); err != nil || token.AccessToken == "" {
		return Profile{}, ErrDenied
	}

	infoURL := c.UserInfoURL
	if infoURL == "" {
		infoURL = "https://openidconnect.googleapis.com/v1/userinfo"
	}
	infoReq, err := http.NewRequestWithContext(ctx, http.MethodGet, infoURL, nil)
	if err != nil {
		return Profile{}, err
	}
	infoReq.Header.Set("Authorization", "Bearer "+token.AccessToken)

	infoBody, err := do(c.HTTP, infoReq)
	if err != nil {
		return Profile{}, err
	}

	var info struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		EmailVerified any    `json:"email_verified"`
	}
	if err := json.Unmarshal(infoBody, &info); err != nil || info.Sub == "" {
		return Profile{}, ErrDenied
	}

	return Profile{
		Subject:  info.Sub,
		Email:    info.Email,
		Name:     info.Name,
		Verified: emailVerified(info.EmailVerified),
	}, nil
}

// DisplayName picks a stored display name from Google profile or email local-part.
func DisplayName(name, email string, maxRunes int) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name, _, _ = strings.Cut(email, "@")
	}
	if utf8.RuneCountInString(name) > maxRunes {
		name = string([]rune(name)[:maxRunes])
	}
	if name == "" {
		return "WiChat"
	}

	return name
}

func do(client *http.Client, req *http.Request) ([]byte, error) {
	res, err := client.Do(req)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, ErrUnavailable
	}
	if res.StatusCode != http.StatusOK {
		return nil, ErrDenied
	}

	return body, nil
}

func emailVerified(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(typed, "true")
	default:
		return false
	}
}
