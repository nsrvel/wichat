package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
)

func TestRedisDevOTPStoreRoundTrip(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := repository.NewRedisDevOTPStore(rdb)
	ctx := context.Background()

	if err := store.Put(ctx, "ch-1", "ab12cd", time.Minute); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(ctx, "ch-1")
	if err != nil || got != "AB12CD" {
		t.Fatalf("got %q err %v", got, err)
	}
}
