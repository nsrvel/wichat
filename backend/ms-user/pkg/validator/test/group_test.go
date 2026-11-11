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

	desc, err := validator.GroupDescription("  hello  ")
	if err != nil || desc != "hello" {
		t.Fatalf("description %q %v", desc, err)
	}
	if _, err := validator.GroupDescription(strings.Repeat("a", constants.MaxGroupDescriptionRunes+1)); err == nil {
		t.Fatal("long description")
	}

	color, err := validator.GroupBannerColor("volcano")
	if err != nil || color != "volcano" {
		t.Fatalf("banner color %q %v", color, err)
	}
	if _, err := validator.GroupBannerColor("not-a-color"); err == nil {
		t.Fatal("bad banner color")
	}

	if err := validator.GroupBannerUpload("image/png", 1024); err != nil {
		t.Fatal(err)
	}
	if err := validator.GroupBannerUpload("image/png", constants.MaxGroupBannerBytes+1); err == nil {
		t.Fatal("banner too large")
	}
}
