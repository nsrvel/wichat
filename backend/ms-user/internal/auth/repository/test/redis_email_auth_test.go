package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestEmailAuthRegisterPendingRoundTrip(t *testing.T) {
	store := newEmailAuthStore(t)
	ctx := context.Background()
	ttl := 10 * time.Minute
	pending := models.RegisterPending{
		Email:        "new@example.com",
		PasswordHash: "hash",
		DisplayName:  "New",
		CodeHMAC:     authRedis.OTPCodeHMAC(testOTPSecret, "654321"),
	}

	if err := store.PutRegisterPending(ctx, "ch-1", pending, ttl); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetRegisterPending(ctx, "ch-1")
	if err != nil || got.Email != pending.Email || got.CodeHMAC != pending.CodeHMAC {
		t.Fatalf("get: %+v %v", got, err)
	}

	rec, err := store.VerifyRegisterCode(ctx, "ch-1", "654321", testOTPSecret, 5)
	if err != nil || rec.Email != pending.Email {
		t.Fatalf("verify: %+v %v", rec, err)
	}
}

func TestEmailAuthRegisterWrongCodeThenExhaust(t *testing.T) {
	store := newEmailAuthStore(t)
	ctx := context.Background()

	pending := models.RegisterPending{
		Email:    "new@example.com",
		CodeHMAC: authRedis.OTPCodeHMAC(testOTPSecret, "111111"),
	}
	if err := store.PutRegisterPending(ctx, "ch-2", pending, 10*time.Minute); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		_, err := store.VerifyRegisterCode(ctx, "ch-2", "000000", testOTPSecret, 5)
		if !errors.Is(err, constants.ErrNotFound) {
			t.Fatalf("wrong %d: %v", i+1, err)
		}
	}

	_, err := store.VerifyRegisterCode(ctx, "ch-2", "111111", testOTPSecret, 5)
	if !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("expected deleted challenge, got %v", err)
	}
}

func TestEmailAuthLoginChallengeVerify(t *testing.T) {
	store := newEmailAuthStore(t)
	ctx := context.Background()

	pending := models.LoginPending{
		UserID:   "user-1",
		Email:    "ada@example.com",
		CodeHMAC: authRedis.OTPCodeHMAC(testOTPSecret, "222222"),
	}
	if err := store.PutLoginChallenge(ctx, "login-ch", pending, 10*time.Minute); err != nil {
		t.Fatal(err)
	}

	got, err := store.VerifyLoginCode(ctx, "login-ch", "222222", testOTPSecret, 5)
	if err != nil || got.UserID != "user-1" {
		t.Fatalf("verify login: %+v %v", got, err)
	}
}

func TestEmailAuthResendCooldown(t *testing.T) {
	store := newEmailAuthStore(t)
	ctx := context.Background()
	email := "ada@example.com"

	if err := store.EnsureCanSendEmailCode(ctx, email); err != nil {
		t.Fatal(err)
	}

	wait, err := store.RecordEmailCodeSent(ctx, email)
	if err != nil || wait <= 0 {
		t.Fatalf("record sent: %v %v", wait, err)
	}

	err = store.EnsureCanSendEmailCode(ctx, email)
	var rl *models.ResetRateLimitError
	if !errors.As(err, &rl) || rl.RetryAfter <= 0 {
		t.Fatalf("expected cooldown, got %v", err)
	}

	if err := store.ClearEmailCodePenalty(ctx, email); err != nil {
		t.Fatal(err)
	}
	if err := store.EnsureCanSendEmailCode(ctx, email); err != nil {
		t.Fatalf("after clear: %v", err)
	}
}

func TestEmailAuthUpdateRegisterCode(t *testing.T) {
	store := newEmailAuthStore(t)
	ctx := context.Background()
	ttl := 10 * time.Minute

	pending := models.RegisterPending{
		Email:    "u@example.com",
		CodeHMAC: authRedis.OTPCodeHMAC(testOTPSecret, "111111"),
		Attempts: 2,
	}
	if err := store.PutRegisterPending(ctx, "ch-up", pending, ttl); err != nil {
		t.Fatal(err)
	}

	newHMAC := authRedis.OTPCodeHMAC(testOTPSecret, "999999")
	if err := store.UpdateRegisterCode(ctx, "ch-up", newHMAC, ttl); err != nil {
		t.Fatal(err)
	}

	_, err := store.VerifyRegisterCode(ctx, "ch-up", "111111", testOTPSecret, 5)
	if !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("old code should fail: %v", err)
	}

	rec, err := store.VerifyRegisterCode(ctx, "ch-up", "999999", testOTPSecret, 5)
	if err != nil || rec.Email != pending.Email {
		t.Fatalf("new code: %+v %v", rec, err)
	}
}

func TestEmailAuthDeleteRegisterPending(t *testing.T) {
	store := newEmailAuthStore(t)
	ctx := context.Background()

	if err := store.PutRegisterPending(ctx, "ch-del", models.RegisterPending{Email: "x@y.com"}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteRegisterPending(ctx, "ch-del"); err != nil {
		t.Fatal(err)
	}

	_, err := store.GetRegisterPending(ctx, "ch-del")
	if !errors.Is(err, constants.ErrNotFound) {
		t.Fatalf("get after delete: %v", err)
	}
}
