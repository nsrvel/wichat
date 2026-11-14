package validator_test

import (
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
)

func TestUserAvatarUpload(t *testing.T) {
	if err := validator.UserAvatarUpload("image/png", 1024); err != nil {
		t.Fatalf("expected valid png upload, got %v", err)
	}
	if err := validator.UserAvatarUpload("text/plain", 1024); err == nil {
		t.Fatal("expected invalid content type error")
	}
	if err := validator.UserAvatarUpload("image/png", constants.MaxUserAvatarBytes+1); err == nil {
		t.Fatal("expected oversized avatar error")
	}
}

func TestUserProfileInput(t *testing.T) {
	dn, p, b, c, err := validator.UserProfileInput("  Garam  ", " he/him ", " Building WiChat ", " #2b2d31 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dn != "Garam" || p != "he/him" || b != "Building WiChat" || c != "#2B2D31" {
		t.Fatalf("unexpected normalized values: %q %q %q %q", dn, p, b, c)
	}

	if _, _, _, _, err := validator.UserProfileInput("A", "", "", ""); err == nil {
		t.Fatal("expected short display name error")
	}
	if _, _, _, _, err := validator.UserProfileInput("Garam", strings.Repeat("x", constants.MaxUserPronounsRunes+1), "", ""); err == nil {
		t.Fatal("expected long pronouns error")
	}
	if _, _, _, _, err := validator.UserProfileInput("Garam", "", strings.Repeat("x", constants.MaxUserBioRunes+1), ""); err == nil {
		t.Fatal("expected long bio error")
	}
	if _, _, _, _, err := validator.UserProfileInput("Garam", "", "", "red"); err == nil {
		t.Fatal("expected invalid hex banner_color error")
	}
}
