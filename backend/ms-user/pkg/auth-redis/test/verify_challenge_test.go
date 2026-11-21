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

func TestVerifyChallengeJSONWrongThenSuccess(t *testing.T) {
	t.Parallel()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	key := "wichat:test:ch:1"
	secret := []byte("test-secret-32-bytes-minimum!!")
	hmac := authRedis.OTPCodeHMAC(secret, "ABC123")
	payload, _ := json.Marshal(map[string]any{
		"code_hmac": hmac,
		"attempts":  0,
		"user_id":   "u1",
	})
	if err := mr.Set(key, string(payload)); err != nil {
		t.Fatal(err)
	}
	mr.SetTTL(key, 10*time.Minute)

	_, err := authRedis.VerifyChallengeJSON(ctx, rdb, key, authRedis.HMACCode(secret, "WRONG1"), 5)
	if !authRedis.IsVerifyNotFound(err) {
		t.Fatalf("wrong code: %v", err)
	}

	got, err := authRedis.VerifyChallengeJSON(ctx, rdb, key, hmac, 5)
	if err != nil {
		t.Fatal(err)
	}
	var rec map[string]any
	if err := authRedis.DecodeJSON(got, &rec); err != nil || rec["user_id"] != "u1" {
		t.Fatalf("payload: %v %v", rec, err)
	}
	if mr.Exists(key) {
		t.Fatal("key should be deleted on success")
	}
}

func TestVerifyChallengeJSONConcurrentSingleWinner(t *testing.T) {
	t.Parallel()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	key := "wichat:test:ch:race"
	secret := []byte("test-secret-32-bytes-minimum!!")
	hmac := authRedis.OTPCodeHMAC(secret, "WIN123")
	payload, _ := json.Marshal(map[string]any{"code_hmac": hmac, "attempts": 0})
	if err := mr.Set(key, string(payload)); err != nil {
		t.Fatal(err)
	}
	mr.SetTTL(key, 10*time.Minute)

	var wg sync.WaitGroup
	winners := make(chan struct{}, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := authRedis.VerifyChallengeJSON(ctx, rdb, key, hmac, 5); err == nil {
				winners <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(winners)
	n := 0
	for range winners {
		n++
	}
	if n != 1 {
		t.Fatalf("expected exactly 1 winner, got %d", n)
	}
}

func TestVerifyChallengeExhaustAttempts(t *testing.T) {
	t.Parallel()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	key := "wichat:test:ch:ex"
	secret := []byte("test-secret-32-bytes-minimum!!")
	hmac := authRedis.OTPCodeHMAC(secret, "GOOD99")
	payload, _ := json.Marshal(map[string]any{"code_hmac": hmac, "attempts": 0})
	if err := mr.Set(key, string(payload)); err != nil {
		t.Fatal(err)
	}

	bad := authRedis.HMACCode(secret, "BAD000")
	for i := 0; i < 5; i++ {
		_, err := authRedis.VerifyChallengeJSON(ctx, rdb, key, bad, 5)
		if !errorsIsNotFound(err) {
			t.Fatalf("attempt %d: %v", i+1, err)
		}
	}
	_, err := authRedis.VerifyChallengeJSON(ctx, rdb, key, hmac, 5)
	if !errorsIsNotFound(err) {
		t.Fatalf("challenge should be gone: %v", err)
	}
}

func errorsIsNotFound(err error) bool {
	return authRedis.IsVerifyNotFound(err) || errors.Is(err, constants.ErrNotFound)
}
