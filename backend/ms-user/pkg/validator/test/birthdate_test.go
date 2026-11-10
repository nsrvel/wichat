package validator_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
)

func TestBirthDateAndAge(t *testing.T) {
	t.Parallel()

	_, err := validator.ParseBirthDate("")
	if err == nil {
		t.Fatal("empty birth date")
	}
	birth, err := validator.ParseBirthDate("2000-01-15")
	if err != nil || birth == nil {
		t.Fatal(err)
	}
	if !validator.MeetsMinAge(*birth) {
		t.Fatal("adult birth should meet min age")
	}
	years := validator.AgeYears(*birth)
	if years < 20 {
		t.Fatalf("age %d", years)
	}
}
