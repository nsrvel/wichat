package authRedis_test

import (
	"testing"

	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
)

func TestEmailKeyStable(t *testing.T) {
	t.Parallel()

	k1 := authRedis.EmailKey("Ada@Example.com")
	k2 := authRedis.EmailKey(" ada@example.com ")
	if k1 != k2 || k1 == "" {
		t.Fatalf("keys %q %q", k1, k2)
	}
}
