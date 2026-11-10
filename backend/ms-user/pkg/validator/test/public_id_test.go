package validator_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
)

func TestGroupPublicID(t *testing.T) {
	t.Parallel()

	_, err := validator.GroupPublicID("")
	if err == nil {
		t.Fatal("expected error")
	}

	got, err := validator.GroupPublicID("  a1b2c3d4e5f6789012345678  ")
	if err != nil {
		t.Fatal(err)
	}
	if got != "a1b2c3d4e5f6789012345678" {
		t.Fatalf("got %q", got)
	}
}
