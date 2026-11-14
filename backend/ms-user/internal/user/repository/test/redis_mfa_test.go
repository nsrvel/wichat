package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestUserMFAEnrollmentStoreLifecycle(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	store := repository.NewMFAEnrollmentStore(rdb)
	ctx := context.Background()

	if err := store.PutTOTPSetup(ctx, "st-1", models.TOTPSetupPending{UserID: "u-1", Secret: "SEC"}, 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	gotTOTP, err := store.TakeTOTPSetup(ctx, "st-1")
	if err != nil || gotTOTP.UserID != "u-1" || gotTOTP.Secret != "SEC" {
		t.Fatalf("TakeTOTPSetup() = %+v, %v", gotTOTP, err)
	}
	if _, err := store.TakeTOTPSetup(ctx, "st-1"); !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("expected single-use TakeTOTPSetup ErrNotFound, got %v", err)
	}

	if err := store.PutWebAuthnRegister(ctx, "wc-1", models.WebAuthnRegisterPending{UserID: "u-1", KeyName: "Key", SessionJSON: "{}"}, 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	gotWA, err := store.TakeWebAuthnRegister(ctx, "wc-1")
	if err != nil || gotWA.KeyName != "Key" {
		t.Fatalf("TakeWebAuthnRegister() = %+v, %v", gotWA, err)
	}
	if _, err := store.TakeWebAuthnRegister(ctx, "wc-1"); !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("expected single-use TakeWebAuthnRegister ErrNotFound, got %v", err)
	}
}
