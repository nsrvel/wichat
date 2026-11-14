package mfa_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

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

	var zeroKey [32]byte
	if _, err := mfa.DecryptSecret(zeroKey, make([]byte, 32)); err == nil {
		t.Fatal("expected invalid ciphertext error")
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
