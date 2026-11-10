package validator_test

import (
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
)

func TestGroupNameAndIcon(t *testing.T) {
	t.Parallel()

	gname, err := validator.GroupName("  Acme  ")
	if err != nil || gname != "Acme" {
		t.Fatalf("%q %v", gname, err)
	}
	if _, err := validator.GroupName(strings.Repeat("a", constants.MaxGroupNameRunes+1)); err == nil {
		t.Fatal("long group name")
	}

	if err := validator.GroupIconUpload("image/png", 1024); err != nil {
		t.Fatal(err)
	}
	if err := validator.GroupIconUpload("text/plain", 1024); err == nil {
		t.Fatal("bad mime")
	}
}
