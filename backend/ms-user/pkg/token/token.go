// Package token generates opaque session ids and refresh secrets stored as SHA-256 hashes.
package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// NewRefresh returns a URL-safe plaintext refresh token and its SHA-256 digest for storage.
func NewRefresh() (plain string, hash []byte, err error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", nil, err
	}

	plain = base64.RawURLEncoding.EncodeToString(raw[:])

	return plain, HashRefresh(plain), nil
}

// HashRefresh digests the plaintext refresh token for repository lookup.
func HashRefresh(plain string) []byte {
	sum := sha256.Sum256([]byte(plain))
	out := make([]byte, len(sum))
	copy(out, sum[:])

	return out
}

// NewID returns a random UUID-like string for users, sessions, and refresh rows.
func NewID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
