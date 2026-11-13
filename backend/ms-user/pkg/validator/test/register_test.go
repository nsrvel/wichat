package validator_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
)

const validPassword = "correct-horse-battery"

func TestRegisterInput(t *testing.T) {
	t.Parallel()

	_, _, _, _, err := validator.RegisterInput(models.RegisterInput{
		Email: "bad", Username: "ada.lovelace", Password: validPassword, DisplayName: "Ada", BirthDate: "2000-01-15",
	})
	if err == nil {
		t.Fatal("expected invalid email")
	}

	_, _, _, _, err = validator.RegisterInput(models.RegisterInput{
		Email: "ada@example.com", Username: "ab", Password: validPassword, DisplayName: "Ada", BirthDate: "2000-01-15",
	})
	if err == nil {
		t.Fatal("expected invalid username")
	}

	email, handle, name, birth, err := validator.RegisterInput(models.RegisterInput{
		Email: "ada@example.com", Username: "Ada.Lovelace", Password: validPassword, DisplayName: "Ada", BirthDate: "2000-01-15",
	})
	if err != nil || email != "ada@example.com" || handle != "ada.lovelace" || name != "Ada" || birth == nil {
		t.Fatalf("%q %q %q %+v %v", email, handle, name, birth, err)
	}
}
