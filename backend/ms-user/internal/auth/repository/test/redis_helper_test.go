package repository_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
)

var testOTPSecret = []byte("test-secret-must-be-32-characters!!")

func newMiniRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)

	return redis.NewClient(&redis.Options{Addr: mr.Addr()}), mr
}

func newPasswordResetStore(t *testing.T) authrepo.PasswordResetStore {
	rdb, _ := newMiniRedis(t)

	return authrepo.NewRedisPasswordResetStore(rdb)
}

func newEmailAuthStore(t *testing.T) authrepo.EmailAuthStore {
	rdb, _ := newMiniRedis(t)

	return authrepo.NewRedisEmailAuthStore(rdb)
}
