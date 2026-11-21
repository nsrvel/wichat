package authRedis_test

import (
	"testing"

	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
)

func TestDevE2EOTPKey(t *testing.T) {
	const id = "challenge-abc"
	got := authRedis.DevE2EOTPKey(id)
	want := "ms-user:dev:e2e-otp:challenge-abc"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
