package token

import (
	"regexp"
	"testing"
)

func TestNewInviteTokenLength(t *testing.T) {
	tok, err := NewInviteToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(tok) != 8 {
		t.Fatalf("invite token length %d, want 8", len(tok))
	}
}

func TestNewPublicIDFormat(t *testing.T) {
	id, err := NewPublicID()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{24}$`).MatchString(id) {
		t.Fatalf("id %q", id)
	}
}
