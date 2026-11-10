package validator_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
)

func TestChannelPublicID(t *testing.T) {
	t.Parallel()

	_, err := validator.ChannelPublicID("bad")
	if err == nil {
		t.Fatal("expected error")
	}
	got, err := validator.ChannelPublicID("  A1B2C3D4E5F6789012345678  ")
	if err != nil || got != "a1b2c3d4e5f6789012345678" {
		t.Fatalf("got %q %v", got, err)
	}
}
