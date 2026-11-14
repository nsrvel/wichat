package mfa

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"strings"
)

// Unambiguous Crockford-like alphabet (no 0/O, 1/I/L).
const recoveryAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// GenerateRecoveryCodes returns count cryptographically random XXXX-XXXX backup codes.
func GenerateRecoveryCodes(count int) ([]string, error) {
	if count <= 0 {
		return nil, nil
	}
	maxIdx := big.NewInt(int64(len(recoveryAlphabet)))
	codes := make([]string, count)
	for i := range count {
		var b strings.Builder
		b.Grow(9)
		for pos := range 8 {
			if pos == 4 {
				b.WriteByte('-')
			}
			n, err := rand.Int(rand.Reader, maxIdx)
			if err != nil {
				return nil, err
			}
			b.WriteByte(recoveryAlphabet[n.Int64()])
		}
		codes[i] = b.String()
	}

	return codes, nil
}

// NormalizeRecoveryCode strips whitespace/hyphens and uppercases a user-entered backup code.
func NormalizeRecoveryCode(raw string) string {
	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range strings.ToUpper(strings.TrimSpace(raw)) {
		if r == '-' || r == ' ' {
			continue
		}
		b.WriteRune(r)
	}

	return b.String()
}

// HashRecoveryCode computes a server-keyed HMAC-SHA256 digest for storing/comparing a recovery code.
func HashRecoveryCode(secret []byte, raw string) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte("wichat:mfa:recovery:v1:"))
	mac.Write([]byte(NormalizeRecoveryCode(raw)))
	return mac.Sum(nil)
}
