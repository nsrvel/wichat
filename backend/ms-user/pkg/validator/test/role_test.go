package validator_test

import (
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
)

func TestRoleNameAndColor(t *testing.T) {
	t.Parallel()

	if _, err := validator.RoleName("   "); err == nil {
		t.Fatal("expected empty role name error")
	}
	if _, err := validator.RoleName(strings.Repeat("a", constants.MaxRoleNameRunes+1)); err == nil {
		t.Fatal("expected max role name error")
	}
	if _, err := validator.RoleName("Mod\nTeam"); err == nil {
		t.Fatal("expected control char error")
	}
	got, err := validator.RoleName("  Moderator  ")
	if err != nil || got != "Moderator" {
		t.Fatalf("got %q %v", got, err)
	}

	if color, err := validator.RoleColor(""); err != nil || color != "" {
		t.Fatalf("empty color: %q %v", color, err)
	}
	if color, err := validator.RoleColor("#5865F2"); err != nil || color != "#5865F2" {
		t.Fatalf("valid hex: %q %v", color, err)
	}
	if _, err := validator.RoleColor("red"); err == nil {
		t.Fatal("expected invalid hex color")
	}
}
