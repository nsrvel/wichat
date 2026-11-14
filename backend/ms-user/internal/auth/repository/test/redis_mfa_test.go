package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestAuthMFAStoreLifecycle(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	store := repository.NewMFAStore(rdb)
	ctx := context.Background()

	rec := models.LoginMFAPending{
		UserID:  "u-1",
		Email:   "a@x.com",
		Methods: []string{"totp", "recovery"},
	}
	if err := store.PutLoginMFA(ctx, "ch-1", rec, 5*time.Minute); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetLoginMFA(ctx, "ch-1")
	if err != nil || got.UserID != "u-1" || len(got.Methods) != 2 {
		t.Fatalf("GetLoginMFA() = %+v, %v", got, err)
	}

	got.Attempts = 2
	if err := store.UpdateLoginMFA(ctx, "ch-1", got, 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	got2, err := store.GetLoginMFA(ctx, "ch-1")
	if err != nil || got2.Attempts != 2 {
		t.Fatalf("after update: %+v, %v", got2, err)
	}

	if err := store.DeleteLoginMFA(ctx, "ch-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetLoginMFA(ctx, "ch-1"); !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}
