package password_test

import (
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/password"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

const testPass = "correct-horse-battery"

func TestHashVerifyRoundTrip(t *testing.T) {
	t.Parallel()

	encoded, err := password.Hash(testPass)
	if err != nil || encoded == "" {
		t.Fatalf("hash: %q %v", encoded, err)
	}
	if !password.Verify(encoded, testPass) {
		t.Fatal("verify failed")
	}
	if password.Verify(encoded, "wrong") {
		t.Fatal("wrong password accepted")
	}
}

func TestValidLengthAndValidate(t *testing.T) {
	t.Parallel()

	short := string(make([]byte, constants.MinPasswordLen-1))
	if password.ValidLength(short) {
		t.Fatal("short accepted")
	}
	if err := password.Validate(short); err == nil {
		t.Fatal("expected validate error")
	}
	if err := password.Validate(short); err == nil {
		t.Fatal("expected validate error")
	} else if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindInvalidInput {
		t.Fatalf("err %v", err)
	}

	if !password.ValidLength(testPass) || password.Validate(testPass) != nil {
		t.Fatal("valid password rejected")
	}
}

func TestVerifyRejectsMalformedHash(t *testing.T) {
	t.Parallel()

	if password.Verify("not-a-hash", testPass) {
		t.Fatal("malformed accepted")
	}
}

func TestDummyHashStable(t *testing.T) {
	t.Parallel()

	a, err := password.DummyHash()
	if err != nil {
		t.Fatal(err)
	}
	b, err := password.DummyHash()
	if err != nil || a != b {
		t.Fatalf("dummy unstable: %q %q", a, b)
	}
}

func TestValidateRejectsBounds(t *testing.T) {
	t.Parallel()

	if err := password.Validate("short"); err == nil {
		t.Fatal("too short")
	}
	long := strings.Repeat("a", 129)
	if err := password.Validate(long); err == nil {
		t.Fatal("too long")
	}
	if err := password.Validate(strings.Repeat("a", 12)); err != nil {
		t.Fatal(err)
	}
}
