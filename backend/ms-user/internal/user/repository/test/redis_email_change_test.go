package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestRedisEmailChangeStoreFlow(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	store := userrepo.NewRedisEmailChangeStore(rdb)
	ctx := context.Background()
	secret := []byte("01234567890123456789012345678901")

	rec := models.EmailChangePending{
		UserID:   "u-1",
		Email:    "a@example.com",
		CodeHMAC: authRedis.OTPCodeHMAC(secret, "ABC123"),
	}
	if err := store.PutChallenge(ctx, "ch-1", rec, 5*time.Minute); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetChallenge(ctx, "ch-1")
	if err != nil || got.UserID != "u-1" {
		t.Fatalf("unexpected get challenge: %+v %v", got, err)
	}

	if _, err := store.VerifyChallenge(ctx, "ch-1", "WRONG1", secret, 2); !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("expected not found on wrong code, got %v", err)
	}

	if _, err := store.VerifyChallenge(ctx, "ch-1", "WRONG2", secret, 2); !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("expected not found on max attempts, got %v", err)
	}

	if _, err := store.GetChallenge(ctx, "ch-1"); !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("expected challenge deleted after max attempts, got %v", err)
	}

	if err := store.PutChallenge(ctx, "ch-2", rec, 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	verified, err := store.VerifyChallenge(ctx, "ch-2", "abc123", secret, 5)
	if err != nil || verified.Email != "a@example.com" {
		t.Fatalf("expected verify success, got %+v %v", verified, err)
	}
	if err := store.DeleteChallenge(ctx, "ch-2"); err != nil {
		t.Fatal(err)
	}

	if err := store.SaveChangeToken(ctx, "tok-1", "u-1", 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	uid, err := store.ConsumeChangeToken(ctx, "tok-1")
	if err != nil || uid != "u-1" {
		t.Fatalf("expected consumed user id u-1, got %q %v", uid, err)
	}
	if _, err := store.ConsumeChangeToken(ctx, "tok-1"); !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("expected second consume to fail, got %v", err)
	}
}
