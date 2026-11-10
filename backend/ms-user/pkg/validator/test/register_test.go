package validator_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
)

const validPassword = "correct-horse-battery"

func TestRegisterInput(t *testing.T) {
	t.Parallel()

	_, _, _, err := validator.RegisterInput(models.RegisterInput{
		Email: "bad", Password: validPassword, DisplayName: "Ada", BirthDate: "2000-01-15",
	})
	if err == nil {
		t.Fatal("expected invalid email")
	}

	email, name, birth, err := validator.RegisterInput(models.RegisterInput{
		Email: "ada@example.com", Password: validPassword, DisplayName: "Ada", BirthDate: "2000-01-15",
	})
	if err != nil || email != "ada@example.com" || name != "Ada" || birth == nil {
		t.Fatalf("%q %q %+v %v", email, name, birth, err)
	}
}
