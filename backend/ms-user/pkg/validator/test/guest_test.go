package validator_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
)

func TestGuestRegisterValidation(t *testing.T) {
	t.Parallel()

	_, _, err := validator.GuestRegister(models.GuestRegisterInput{})
	if err == nil {
		t.Fatal("empty name")
	}

	_, _, err = validator.GuestRegister(models.GuestRegisterInput{
		DisplayName: "Kid", BirthDate: "2020-01-01",
	})
	if err == nil {
		t.Fatal("underage")
	}

	name, birth, err := validator.GuestRegister(models.GuestRegisterInput{
		DisplayName: "Visitor", BirthDate: "2000-01-15",
	})
	if err != nil || name != "Visitor" || birth == nil {
		t.Fatalf("%q %+v %v", name, birth, err)
	}
}
