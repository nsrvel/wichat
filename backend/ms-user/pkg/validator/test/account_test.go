package validator_test

import (
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

func TestNewAccount(t *testing.T) {
	t.Parallel()

	email, name, err := validator.NewAccount("ada@example.com", validPassword, " Ada ")
	if err != nil || email != "ada@example.com" || name != "Ada" {
		t.Fatalf("%q %q %v", email, name, err)
	}
}

func TestNewAccountRejectsWeakPassword(t *testing.T) {
	t.Parallel()

	_, _, err := validator.NewAccount("ada@example.com", "short", "Ada")
	if err == nil {
		t.Fatal("expected weak password error")
	}
}

func TestNewAccountRejectsBlankDisplayName(t *testing.T) {
	t.Parallel()

	_, _, err := validator.NewAccount("ada@example.com", validPassword, "   ")
	if err == nil {
		t.Fatal("expected display name error")
	}
	ex, ok := exception.As(err)
	if !ok || ex.Kind != exception.KindInvalidInput {
		t.Fatalf("%v", err)
	}
}

func TestNewAccountRejectsLongDisplayName(t *testing.T) {
	t.Parallel()

	_, _, err := validator.NewAccount("ada@example.com", validPassword, strings.Repeat("字", constants.MaxDisplayNameRunes+1))
	if err == nil {
		t.Fatal("expected long display name error")
	}
}
