package validator_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
)

func TestInstanceSetupLogoURL(t *testing.T) {
	t.Parallel()

	_, err := validator.InstanceSetup(models.CompleteInstanceSetupInput{
		Name: "WiChat", LogoURL: "http://insecure.example/x", RegistrationMode: constants.RegistrationOpen,
	})
	if err == nil {
		t.Fatal("expected https")
	}

	got, err := validator.InstanceSetup(models.CompleteInstanceSetupInput{
		Name: "WiChat", LogoURL: "https://cdn.example/x.png", RegistrationMode: constants.RegistrationOpen,
	})
	if err != nil || got.GroupLogoURL == "" {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestInstanceSetupHappyPath(t *testing.T) {
	t.Parallel()

	payload, err := validator.InstanceSetup(models.CompleteInstanceSetupInput{
		Name:             "WiChat",
		LogoURL:          "https://cdn.example/logo.png",
		RegistrationMode: constants.RegistrationOpen,
	})
	if err != nil || payload.GroupName != "WiChat" {
		t.Fatalf("%+v %v", payload, err)
	}
}

func TestInstanceSetupInvalidMode(t *testing.T) {
	t.Parallel()

	_, err := validator.InstanceSetup(models.CompleteInstanceSetupInput{
		Name: "WiChat", RegistrationMode: "bogus",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
