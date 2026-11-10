package validator_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()

	got, err := validator.NormalizeEmail("  User@Example.COM ")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != "user@example.com" {
		t.Fatalf("got %q", got)
	}

	_, err = validator.NormalizeEmail("not-an-email")
	if err == nil {
		t.Fatal("expected error")
	}
	if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindInvalidInput {
		t.Fatalf("got %v", err)
	}
}
