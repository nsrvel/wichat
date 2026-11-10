package exception_test

import (
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/exception"
)

func TestValidMessageCodePrefixes(t *testing.T) {
	t.Parallel()

	if !exception.ValidMessageCode(exception.MsgForbidden) {
		t.Fatal("auth prefix")
	}
	if exception.ValidMessageCode("") {
		t.Fatal("empty")
	}
	if exception.ValidMessageCode("nope") {
		t.Fatal("unknown")
	}
}
