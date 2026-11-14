package mfa

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const totpPeriodSeconds = 30

// DeriveEncryptionKey derives a 32-byte AES-256 key with domain separation from an app secret.
func DeriveEncryptionKey(secret string) [32]byte {
	return sha256.Sum256([]byte("wichat:mfa:totp:v1:" + secret))
}

// EncryptSecret seals a TOTP secret using AES-256-GCM with a random 12-byte nonce.
func EncryptSecret(key [32]byte, plaintext string) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

// DecryptSecret opens an AES-256-GCM sealed TOTP secret.
func DecryptSecret(key [32]byte, ciphertext []byte) (string, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, body := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plain, err := gcm.Open(nil, nonce, body, nil)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}

// GenerateTOTP creates a new RFC 6238 TOTP key and otpauth URI for an authenticator app.
func GenerateTOTP(issuer, accountName string) (secret, otpauthURI string, err error) {
	if strings.TrimSpace(issuer) == "" {
		issuer = "WiChat"
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
		Period:      totpPeriodSeconds,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

// VerifyTOTP validates a 6-digit TOTP code within ±1 period skew and enforces monotonic step replay protection.
func VerifyTOTP(secret, code string, now time.Time, lastUsedStep int64) (int64, bool) {
	clean := strings.TrimSpace(code)
	if len(clean) != 6 || secret == "" {
		return 0, false
	}

	baseStep := now.Unix() / totpPeriodSeconds
	for _, offset := range []int64{0, -1, 1} {
		step := baseStep + offset
		if step <= lastUsedStep {
			continue
		}
		t := time.Unix(step*totpPeriodSeconds, 0).UTC()
		ok, err := totp.ValidateCustom(clean, secret, t, totp.ValidateOpts{
			Period:    totpPeriodSeconds,
			Skew:      0,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		})
		if err == nil && ok {
			return step, true
		}
	}

	return 0, false
}
