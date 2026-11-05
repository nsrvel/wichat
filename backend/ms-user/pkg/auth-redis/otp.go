package authRedis

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

const (
	ResendCooldownBase = 60 * time.Second
	ResendCooldownMax  = 15 * time.Minute
	StrikeKeyTTL       = 24 * time.Hour
)

// EmailKey returns a stable Redis key segment for a normalized email address.
func EmailKey(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))

	return hex.EncodeToString(sum[:])
}

// HMACCode hashes code with secret (caller passes normalized OTP when needed).
func HMACCode(secret []byte, code string) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(code))

	return hex.EncodeToString(mac.Sum(nil))
}

// NormalizeOTP uppercases and trims user-entered email OTP input.
func NormalizeOTP(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// OTPCodeHMAC hashes a normalized email OTP for storage.
func OTPCodeHMAC(secret []byte, code string) string {
	return HMACCode(secret, NormalizeOTP(code))
}

// ResendCooldownForStrikes returns exponential backoff from strike count (capped).
func ResendCooldownForStrikes(strikes int) time.Duration {
	if strikes < 1 {
		return 0
	}
	d := ResendCooldownBase
	for i := 1; i < strikes; i++ {
		d *= 2
		if d >= ResendCooldownMax {
			return ResendCooldownMax
		}
	}

	return d
}
