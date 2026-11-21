package authRedis_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestVerifyRegisterChallengeJSONKeepsKeyOnSuccess(t *testing.T) {
	t.Parallel()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	key := "ms-user:register:pending:ch-reg"
	secret := []byte("test-secret-32-bytes-minimum!!")
	hmac := authRedis.OTPCodeHMAC(secret, "REG123")
	payload, _ := json.Marshal(map[string]any{
		"code_hmac": hmac,
		"attempts":  0,
		"email":     "a@example.com",
	})
	if err := mr.Set(key, string(payload)); err != nil {
		t.Fatal(err)
	}
	mr.SetTTL(key, 10*time.Minute)

	got, err := authRedis.VerifyRegisterChallengeJSON(ctx, rdb, key, hmac, 5)
	if err != nil {
		t.Fatal(err)
	}
	var rec map[string]any
	if err := authRedis.DecodeJSON(got, &rec); err != nil || rec["otp_verified"] != true {
		t.Fatalf("verified payload: %v %v", rec, err)
	}
	if !mr.Exists(key) {
		t.Fatal("register pending key should remain after OTP success")
	}

	got2, err := authRedis.VerifyRegisterChallengeJSON(ctx, rdb, key, hmac, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got2) == 0 {
		t.Fatal("expected idempotent verify")
	}
}

func TestVerifyRegisterChallengeJSONConcurrentRetryCreate(t *testing.T) {
	t.Parallel()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	key := "ms-user:register:pending:race-reg"
	secret := []byte("test-secret-32-bytes-minimum!!")
	hmac := authRedis.OTPCodeHMAC(secret, "RACE01")
	payload, _ := json.Marshal(map[string]any{"code_hmac": hmac, "attempts": 0})
	if err := mr.Set(key, string(payload)); err != nil {
		t.Fatal(err)
	}
	mr.SetTTL(key, 10*time.Minute)

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = authRedis.VerifyRegisterChallengeJSON(ctx, rdb, key, hmac, 5)
		}()
	}
	wg.Wait()

	if !mr.Exists(key) {
		t.Fatal("expected pending to survive concurrent verify")
	}
}

func TestVerifyRegisterChallengeJSONWrongCodeIncrements(t *testing.T) {
	t.Parallel()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	key := "ms-user:register:pending:wrong"
	secret := []byte("test-secret-32-bytes-minimum!!")
	hmac := authRedis.OTPCodeHMAC(secret, "GOOD11")
	payload, _ := json.Marshal(map[string]any{"code_hmac": hmac, "attempts": 0})
	if err := mr.Set(key, string(payload)); err != nil {
		t.Fatal(err)
	}
	mr.SetTTL(key, time.Minute)

	_, err := authRedis.VerifyRegisterChallengeJSON(ctx, rdb, key, authRedis.HMACCode(secret, "BAD111"), 5)
	if !authRedis.IsVerifyNotFound(err) {
		t.Fatalf("wrong code: %v", err)
	}
	raw, _ := mr.Get(key)
	var rec map[string]any
	if err := json.Unmarshal([]byte(raw), &rec); err != nil || rec["attempts"] != float64(1) {
		t.Fatalf("attempts: %s %v", raw, err)
	}
}

func TestVerifyRegisterChallengeJSONRejectsWrongCodeAfterVerified(t *testing.T) {
	t.Parallel()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	key := "ms-user:register:pending:verified"
	secret := []byte("test-secret-32-bytes-minimum!!")
	hmac := authRedis.OTPCodeHMAC(secret, "GOOD11")
	payload, _ := json.Marshal(map[string]any{
		"code_hmac":    hmac,
		"otp_verified": true,
		"attempts":     0,
	})
	if err := mr.Set(key, string(payload)); err != nil {
		t.Fatal(err)
	}
	if _, err := authRedis.VerifyRegisterChallengeJSON(ctx, rdb, key, authRedis.HMACCode(secret, "BADBAD"), 5); !authRedis.IsVerifyNotFound(err) {
		t.Fatalf("wrong after verified: %v", err)
	}
	got, err := authRedis.VerifyRegisterChallengeJSON(ctx, rdb, key, hmac, 5)
	if err != nil || len(got) == 0 {
		t.Fatalf("retry correct after verify: %v", err)
	}
}

func TestVerifyRegisterChallengeJSONExhaustDeletes(t *testing.T) {
	t.Parallel()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	key := "ms-user:register:pending:exhaust"
	secret := []byte("test-secret-32-bytes-minimum!!")
	hmac := authRedis.OTPCodeHMAC(secret, "LAST11")
	payload, _ := json.Marshal(map[string]any{"code_hmac": hmac, "attempts": 0})
	if err := mr.Set(key, string(payload)); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		_, err := authRedis.VerifyRegisterChallengeJSON(ctx, rdb, key, authRedis.HMACCode(secret, "BADBAD"), 5)
		if !authRedis.IsVerifyNotFound(err) && !errors.Is(err, constants.ErrNotFound) {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
	if mr.Exists(key) {
		t.Fatal("expected delete after max wrong attempts")
	}
}
