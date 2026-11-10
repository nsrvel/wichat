package token_test

import (
	"regexp"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/token"
)

func TestNewInviteTokenLength(t *testing.T) {
	tok, err := token.NewInviteToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(tok) != 8 {
		t.Fatalf("invite token length %d, want 8", len(tok))
	}
}

func TestNewRefreshAndHash(t *testing.T) {
	t.Parallel()

	plain, hash, err := token.NewRefresh()
	if err != nil || plain == "" || len(hash) != 32 {
		t.Fatalf("refresh: %q len=%d %v", plain, len(hash), err)
	}
	if string(hash) != string(token.HashRefresh(plain)) {
		t.Fatal("hash mismatch")
	}
}

func TestNewIDShape(t *testing.T) {
	t.Parallel()

	id, err := token.NewID()
	if err != nil || len(id) != 36 {
		t.Fatalf("id %q %v", id, err)
	}
}

func TestNewPublicIDFormat(t *testing.T) {
	id, err := token.NewPublicID()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{24}$`).MatchString(id) {
		t.Fatalf("id %q", id)
	}
}
