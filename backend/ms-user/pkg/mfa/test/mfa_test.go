package mfa_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pquerna/otp/totp"
	"github.com/wichat/wichat/backend/ms-user/pkg/mfa"
)

func TestEncryptDecryptTOTPSecret(t *testing.T) {
	key := mfa.DeriveEncryptionKey("super-secret-jwt-key-32-chars!!")
	sealed, err := mfa.EncryptSecret(key, "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(sealed, []byte("JBSWY3DPEHPK3PXP")) {
		t.Fatal("expected sealed ciphertext not to contain plaintext")
	}

	got, err := mfa.DecryptSecret(key, sealed)
	if err != nil || got != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("DecryptSecret() = %q, %v", got, err)
	}

	if _, err := mfa.DecryptSecret(key, []byte("short")); err == nil {
		t.Fatal("expected error for short ciphertext")
	}
}

func TestGenerateAndVerifyTOTPReplayProtection(t *testing.T) {
	secret, uri, err := mfa.GenerateTOTP("", "alice@wichat.dev")
	if err != nil {
		t.Fatal(err)
	}
	if secret == "" || !strings.HasPrefix(uri, "otpauth://totp/") {
		t.Fatalf("unexpected TOTP output: secret=%q uri=%q", secret, uri)
	}

	now := time.Unix(1_700_000_000, 0).UTC()
	code, err := totp.GenerateCode(secret, now)
	if err != nil {
		t.Fatal(err)
	}

	step, ok := mfa.VerifyTOTP(secret, code, now, 0)
	if !ok || step <= 0 {
		t.Fatalf("expected valid TOTP, got step=%d ok=%v", step, ok)
	}

	// Replaying the same step must fail.
	if _, ok := mfa.VerifyTOTP(secret, code, now, step); ok {
		t.Fatal("expected replay on same step to fail")
	}

	// Wrong code must fail.
	if _, ok := mfa.VerifyTOTP(secret, "000000", now, 0); ok {
		if code != "000000" {
			t.Fatal("expected wrong code to fail")
		}
	}
}

func TestRecoveryCodes(t *testing.T) {
	codes, err := mfa.GenerateRecoveryCodes(10)
	if err != nil || len(codes) != 10 {
		t.Fatalf("GenerateRecoveryCodes(10) = %v, %v", codes, err)
	}
	for _, c := range codes {
		if len(c) != 9 || c[4] != '-' {
			t.Fatalf("unexpected recovery code format: %q", c)
		}
	}

	secret := []byte("hmac-secret")
	h1 := mfa.HashRecoveryCode(secret, codes[0])
	h2 := mfa.HashRecoveryCode(secret, strings.ToLower(strings.ReplaceAll(codes[0], "-", " ")))
	if !bytes.Equal(h1, h2) {
		t.Fatal("expected normalized recovery code hashes to match")
	}
}

func TestWebAuthnHelpers(t *testing.T) {
	rpID, origins, name, err := mfa.BuildWebAuthnConfig("http://localhost:8080/", "", "", "")
	if err != nil || rpID != "localhost" || len(origins) != 1 || origins[0] != "http://localhost:8080" || name != "WiChat" {
		t.Fatalf("BuildWebAuthnConfig() = %q, %v, %q, %v", rpID, origins, name, err)
	}

	if _, _, _, err := mfa.BuildWebAuthnConfig("", "", "", ""); err == nil {
		t.Fatal("expected error with empty origins")
	}

	wa, err := mfa.NewWebAuthn("https://wichat.wistudio.my.id", "", "", "WiChat")
	if err != nil || wa == nil {
		t.Fatalf("NewWebAuthn() = %v, %v", wa, err)
	}

	u := mfa.WebAuthnUser{
		ID:          []byte("user-1"),
		Name:        "alice",
		DisplayName: "Alice",
		Credentials: []webauthn.Credential{{
			ID:        []byte("cred-1"),
			Transport: []protocol.AuthenticatorTransport{protocol.USB},
		}},
	}
	if string(u.WebAuthnID()) != "user-1" || u.WebAuthnName() != "alice" || u.WebAuthnDisplayName() != "Alice" || len(u.WebAuthnCredentials()) != 1 {
		t.Fatalf("unexpected WebAuthnUser methods")
	}
	if excl := mfa.CredentialExcludeList(u.Credentials); len(excl) != 1 {
		t.Fatalf("expected 1 exclude descriptor, got %d", len(excl))
	}

	var key [32]byte
	if _, err := mfa.DecryptSecret(key, []byte("short")); err == nil {
		t.Fatal("expected short ciphertext error")
	}
	if _, err := mfa.DecryptSecret(key, make([]byte, 32)); err == nil {
		t.Fatal("expected invalid ciphertext error")
	}
	if _, err := mfa.NewWebAuthn("", "", "", ""); err == nil {
		t.Fatal("expected NewWebAuthn empty origin error")
	}
}
