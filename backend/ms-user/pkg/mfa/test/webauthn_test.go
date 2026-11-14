package mfa_test

import (
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/wichat/wichat/backend/ms-user/pkg/mfa"
)

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

	if _, err := mfa.NewWebAuthn("", "", "", ""); err == nil {
		t.Fatal("expected NewWebAuthn empty origin error")
	}
}
