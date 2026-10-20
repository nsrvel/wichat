package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestPasswordResetOTPVerifySuccess(t *testing.T) {
	store := newPasswordResetStore(t)
	ctx := context.Background()
	email := "ada@example.com"
	ttl := 10 * time.Minute

	if err := store.SaveOTP(ctx, email, "user-1", "123456", testOTPSecret, ttl); err != nil {
		t.Fatal(err)
	}

	userID, err := store.VerifyOTP(ctx, email, "123456", testOTPSecret, 5)
	if err != nil || userID != "user-1" {
		t.Fatalf("verify: %q %v", userID, err)
	}
}

func TestPasswordResetOTPWrongCode(t *testing.T) {
	store := newPasswordResetStore(t)
	ctx := context.Background()
	email := "ada@example.com"

	if err := store.SaveOTP(ctx, email, "user-1", "123456", testOTPSecret, 10*time.Minute); err != nil {
		t.Fatal(err)
	}

	_, err := store.VerifyOTP(ctx, email, "000000", testOTPSecret, 5)
	if !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}

	// Still valid with correct code after one wrong try.
	userID, err := store.VerifyOTP(ctx, email, "123456", testOTPSecret, 5)
	if err != nil || userID != "user-1" {
		t.Fatalf("second verify: %q %v", userID, err)
	}
}

func TestPasswordResetOTPMaxAttempts(t *testing.T) {
	store := newPasswordResetStore(t)
	ctx := context.Background()
	email := "ada@example.com"

	if err := store.SaveOTP(ctx, email, "user-1", "123456", testOTPSecret, 10*time.Minute); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		_, err := store.VerifyOTP(ctx, email, "000000", testOTPSecret, 5)
		if !errors.Is(err, constants.ErrNotFound) {
			t.Fatalf("attempt %d: %v", i+1, err)
		}
	}

	_, err := store.VerifyOTP(ctx, email, "123456", testOTPSecret, 5)
	if !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("expected exhausted challenge, got %v", err)
	}
}

func TestPasswordResetResendCooldown(t *testing.T) {
	store := newPasswordResetStore(t)
	ctx := context.Background()
	email := "ada@example.com"

	if err := store.EnsureCanSendResetCode(ctx, email); err != nil {
		t.Fatal(err)
	}

	wait, err := store.RecordResetCodeSent(ctx, email)
	if err != nil || wait <= 0 {
		t.Fatalf("cooldown: %v %v", wait, err)
	}

	err = store.EnsureCanSendResetCode(ctx, email)
	var rl *models.ResetRateLimitError
	if !errors.As(err, &rl) || rl.RetryAfter <= 0 {
		t.Fatalf("expected rate limit, got %v", err)
	}
}

func TestPasswordResetTokenConsumeOnce(t *testing.T) {
	store := newPasswordResetStore(t)
	ctx := context.Background()

	if err := store.SaveResetToken(ctx, "tok-1", "user-1", "ada@example.com", 15*time.Minute); err != nil {
		t.Fatal(err)
	}

	userID, em, err := store.ConsumeResetToken(ctx, "tok-1")
	if err != nil || userID != "user-1" || em != "ada@example.com" {
		t.Fatalf("consume: %q %q %v", userID, em, err)
	}

	_, _, err = store.ConsumeResetToken(ctx, "tok-1")
	if !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("second consume: %v", err)
	}
}

func TestPasswordResetClearResendPenalty(t *testing.T) {
	store := newPasswordResetStore(t)
	ctx := context.Background()
	email := "ada@example.com"

	if _, err := store.RecordResetCodeSent(ctx, email); err != nil {
		t.Fatal(err)
	}
	if err := store.EnsureCanSendResetCode(ctx, email); err == nil {
		t.Fatal("expected cooldown")
	}

	if err := store.ClearResendPenalty(ctx, email); err != nil {
		t.Fatal(err)
	}
	if err := store.EnsureCanSendResetCode(ctx, email); err != nil {
		t.Fatalf("after clear: %v", err)
	}
}
