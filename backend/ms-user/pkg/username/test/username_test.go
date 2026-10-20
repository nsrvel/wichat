package username_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/username"
)

func TestBaseFromEmail(t *testing.T) {
	got, err := username.BaseFromEmail("ada@example.com")
	if err != nil || got != "ada" {
		t.Fatalf("ada: %q %v", got, err)
	}

	got, err = username.BaseFromEmail("John.Doe+tag@corp.com")
	if err != nil || got != "john.doe" {
		t.Fatalf("john.doe: %q %v", got, err)
	}

	got, err = username.BaseFromEmail("a@x.com")
	if err != nil || got != "user" {
		t.Fatalf("short local: %q %v", got, err)
	}
}
