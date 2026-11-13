package validator_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
)

func TestTextHelpersViaValidators(t *testing.T) {
	t.Parallel()

	if _, err := validator.GroupName("A"); err == nil {
		t.Fatal("expected min group name length")
	}
	if _, err := validator.GroupName("Bad\nName"); err == nil {
		t.Fatal("expected newline rejection in group name")
	}
	if _, err := validator.GroupDescription("Line 1\nLine 2"); err != nil {
		t.Fatalf("expected newline allowed in group description: %v", err)
	}
	if _, err := validator.GroupDescription("Bad\x00Desc"); err == nil {
		t.Fatal("expected null byte rejection in group description")
	}
}
