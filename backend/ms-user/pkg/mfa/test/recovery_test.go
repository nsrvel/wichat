package mfa_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/mfa"
)

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
