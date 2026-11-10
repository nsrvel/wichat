package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestRedisRefreshRotateReuse(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	store := authrepo.NewRedisRefreshStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	ctx := context.Background()
	now := time.Now().UTC()
	expires := now.Add(24 * time.Hour)

	hash1 := []byte("refresh-one")
	rec := models.RefreshToken{ID: "rt-1", SessionID: "sess-1", TokenHash: hash1, CreatedAt: now}
	if err := store.Put(ctx, rec, "user-1", expires); err != nil {
		t.Fatal(err)
	}

	hash2 := []byte("refresh-two")
	next := models.RefreshToken{ID: "rt-2", SessionID: "sess-1", TokenHash: hash2, CreatedAt: now}
	if err := store.Rotate(ctx, hash1, hash2, next, "user-1", expires, now); err != nil {
		t.Fatal(err)
	}

	got, err := store.FindByHash(ctx, hash1)
	if err != nil {
		t.Fatal(err)
	}
	if got.UsedAt == nil {
		t.Fatal("expected used marker for rotated hash")
	}

	err = store.Rotate(ctx, hash1, []byte("refresh-three"), models.RefreshToken{
		ID: "rt-3", SessionID: "sess-1", TokenHash: []byte("refresh-three"), CreatedAt: now,
	}, "user-1", expires, now)
	if !errors.Is(err, constants.ErrAlreadyUsed) {
		t.Fatalf("expected ErrAlreadyUsed, got %v", err)
	}
}

func TestRedisRefreshRevokeForSession(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	store := authrepo.NewRedisRefreshStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	ctx := context.Background()
	now := time.Now().UTC()
	expires := now.Add(24 * time.Hour)
	hash := []byte("refresh-revoke")
	rec := models.RefreshToken{ID: "rt-1", SessionID: "sess-revoke", TokenHash: hash, CreatedAt: now}
	if err := store.Put(ctx, rec, "user-1", expires); err != nil {
		t.Fatal(err)
	}

	if err := store.RevokeForSession(ctx, "sess-revoke"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.FindByHash(ctx, hash); !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("expected not found after revoke, got %v", err)
	}
}

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
