package jwt_test

import (
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-user/pkg/jwt"
)

func TestIssueParseRoundTrip(t *testing.T) {
	t.Parallel()

	secret := []byte("test-secret-must-be-32-characters!!")
	now := time.Now().UTC().Truncate(time.Second)

	raw, exp, err := jwt.Issue(secret, "user-1", "sess-1", now, time.Hour)
	if err != nil || exp != 3600 || raw == "" {
		t.Fatalf("issue: %q %d %v", raw, exp, err)
	}

	uid, sid, err := jwt.Parse(secret, raw)
	if err != nil || uid != "user-1" || sid != "sess-1" {
		t.Fatalf("parse: %q %q %v", uid, sid, err)
	}
}

func TestIssueRejectsHugeTTL(t *testing.T) {
	t.Parallel()

	_, _, err := jwt.Issue([]byte("x"), "u", "s", time.Now(), time.Duration(^uint64(0)/2))
	if err == nil {
		t.Fatal("expected ttl error")
	}
}

func TestParseRejectsTamperedToken(t *testing.T) {
	t.Parallel()

	secret := []byte("test-secret-must-be-32-characters!!")
	raw, _, err := jwt.Issue(secret, "u", "s", time.Now().UTC(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = jwt.Parse(secret, raw+"x")
	if err == nil {
		t.Fatal("expected error")
	}
}
