package authRedis_test

import (
	"testing"
	"time"

	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
)

func TestCooldownSecsBranches(t *testing.T) {
	t.Parallel()

	if authRedis.CooldownSecs(0) != 0 {
		t.Fatal("zero")
	}
	if authRedis.CooldownSecs(500*time.Millisecond) != 1 {
		t.Fatal("sub-second")
	}
	if authRedis.CooldownSecs(3*time.Second) != 3 {
		t.Fatal("seconds")
	}
	if authRedis.CooldownInt32(2_000_000*time.Second) != 1_000_000 {
		t.Fatal("cap")
	}
}
