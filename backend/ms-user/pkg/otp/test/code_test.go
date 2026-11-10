package otp_test

import (
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/otp"
)

func TestGenerateLengthAndAlphabet(t *testing.T) {
	t.Parallel()

	code, err := otp.Generate(6)
	if err != nil || len(code) != 6 {
		t.Fatalf("code %q %v", code, err)
	}
	for _, r := range code {
		if !strings.ContainsRune(constants.OTPAlphabet, r) {
			t.Fatalf("bad rune %q in %q", r, code)
		}
	}
}
