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

func TestChannelName(t *testing.T) {
	t.Parallel()

	if _, err := validator.ChannelName("  "); err == nil {
		t.Fatal("expected empty channel name error")
	}
	if _, err := validator.ChannelName("gen\neral"); err == nil {
		t.Fatal("expected newline error")
	}
	got, err := validator.ChannelName("  general  ")
	if err != nil || got != "general" {
		t.Fatalf("got %q %v", got, err)
	}
}
