package validator_test

import (
	"strings"
	"testing"
	"time"

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

func TestUserStatusNoteInput(t *testing.T) {
	now := time.Date(2026, 11, 15, 3, 0, 0, 0, time.UTC)
	exp := now.Add(time.Hour).Format(time.RFC3339)

	patch, err := validator.UserStatusNoteInput("  In a call  ", " 🎧 ", exp, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if patch.Text != "In a call" || patch.Emoji != "🎧" || patch.ExpiresAt == nil {
		t.Fatalf("unexpected patch: %+v", patch)
	}

	cleared, err := validator.UserStatusNoteInput("   ", "  ", exp, now)
	if err != nil || cleared.Text != "" || cleared.Emoji != "" || cleared.ExpiresAt != nil {
		t.Fatalf("expected cleared patch, got %+v err=%v", cleared, err)
	}

	if _, err := validator.UserStatusNoteInput(strings.Repeat("a", constants.MaxUserStatusNoteRunes+1), "", "", now); err == nil {
		t.Fatal("expected long text error")
	}
	if _, err := validator.UserStatusNoteInput("ok", strings.Repeat("a", constants.MaxUserStatusEmojiRunes+1), "", now); err == nil {
		t.Fatal("expected long emoji error")
	}
	if _, err := validator.UserStatusNoteInput("ok", "", "not-a-date", now); err == nil {
		t.Fatal("expected invalid expires_at error")
	}
	if _, err := validator.UserStatusNoteInput("ok", "", now.Add(-time.Minute).Format(time.RFC3339), now); err == nil {
		t.Fatal("expected past expires_at error")
	}
}

func TestUserPreferencesInput(t *testing.T) {
	patch, err := validator.UserPreferencesInput(" ID ", " 24H ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if patch.Locale != "id" || patch.TimeFormat != "24h" {
		t.Fatalf("unexpected normalized patch: %+v", patch)
	}

	if _, err := validator.UserPreferencesInput("fr", "auto"); err == nil {
		t.Fatal("expected unsupported locale error")
	}
	if _, err := validator.UserPreferencesInput("en", "99h"); err == nil {
		t.Fatal("expected unsupported time_format error")
	}
}
