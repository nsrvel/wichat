package authRedis_test

import (
	"testing"
	"time"

	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
)

func TestOTPCodeHMAC(t *testing.T) {
	t.Parallel()

	secret := []byte("secret")
	h := authRedis.OTPCodeHMAC(secret, " ab12 ")
	if h != authRedis.HMACCode(secret, "AB12") {
		t.Fatal("otp hmac")
	}
}

func TestResendCooldownForStrikes(t *testing.T) {
	t.Parallel()

	if authRedis.ResendCooldownForStrikes(0) != 0 {
		t.Fatal("zero strikes")
	}
	if authRedis.ResendCooldownForStrikes(1) != authRedis.ResendCooldownBase {
		t.Fatal("first strike")
	}
	d := authRedis.ResendCooldownForStrikes(10)
	if d != authRedis.ResendCooldownMax {
		t.Fatalf("max cap %v", d)
	}
	if authRedis.ResendCooldownForStrikes(2) != 2*authRedis.ResendCooldownBase {
		t.Fatalf("double %v", authRedis.ResendCooldownForStrikes(2))
	}
	_ = time.Second
}
