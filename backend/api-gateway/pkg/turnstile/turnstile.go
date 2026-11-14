// Package turnstile verifies optional Cloudflare Turnstile challenge tokens.
package turnstile

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/exception"
)

const defaultVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// Verifier validates Cloudflare Turnstile tokens when configured with a secret key.
type Verifier struct {
	secretKey string
	endpoint  string
	client    *http.Client
}

// NewVerifier creates a Turnstile verifier; empty secretKey disables verification.
func NewVerifier(secretKey string) Verifier {
	return Verifier{
		secretKey: strings.TrimSpace(secretKey),
		endpoint:  defaultVerifyURL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}
}

// NewVerifierWithEndpoint creates a Verifier targeting a custom endpoint (for tests).
func NewVerifierWithEndpoint(secretKey, endpoint string, client *http.Client) Verifier {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return Verifier{
		secretKey: strings.TrimSpace(secretKey),
		endpoint:  strings.TrimSpace(endpoint),
		client:    client,
	}
}

// Enabled reports whether Turnstile verification is active.
func (v Verifier) Enabled() bool {
	return v.secretKey != ""
}

type siteverifyResponse struct {
	Success bool `json:"success"`
}

// Verify checks token against Cloudflare Turnstile; returns nil when disabled.
func (v Verifier) Verify(ctx context.Context, token, remoteIP string) error {
	if !v.Enabled() {
		return nil
	}
	tok := strings.TrimSpace(token)
	if tok == "" {
		return exception.PermissionDenied(exception.MsgCaptchaFailed, nil)
	}

	form := url.Values{}
	form.Set("secret", v.secretKey)
	form.Set("response", tok)
	if ip := strings.TrimSpace(remoteIP); ip != "" {
		form.Set("remoteip", ip)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return exception.Unavailable(exception.MsgUnavailable, nil)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return exception.Unavailable(exception.MsgUnavailable, nil)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return exception.Unavailable(exception.MsgUnavailable, nil)
	}

	var out siteverifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return exception.Unavailable(exception.MsgUnavailable, nil)
	}
	if !out.Success {
		return exception.PermissionDenied(exception.MsgCaptchaFailed, nil)
	}

	return nil
}
