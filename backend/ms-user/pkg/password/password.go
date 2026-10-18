// Package password hashes and verifies credentials with argon2id PHC strings.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 2
	argonMemory  uint32 = 19 * 1024
	argonThreads uint8  = 1
	argonKeyLen         = 32
	argonSaltLen        = 16

	maxArgonMemory  uint32 = 64 * 1024
	maxArgonTime    uint32 = 5
	maxArgonThreads uint8  = 4
)

var (
	dummyOnce sync.Once
	dummyHash string
	dummyErr  error
)

// Hash encodes password as an argon2id PHC string.
func Hash(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	sum := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(sum),
	), nil
}

// Verify reports whether password matches an argon2id PHC string.
func Verify(encoded, password string) bool {
	salt, sum, memory, timeCost, threads, ok := decodeArgon2id(encoded)
	if !ok {
		return false
	}

	got := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, argonKeyLen)

	return subtle.ConstantTimeCompare(got, sum) == 1
}

// DummyHash is a stable hash used when the account is unknown so login timing does not leak existence.
func DummyHash() (string, error) {
	dummyOnce.Do(func() {
		dummyHash, dummyErr = Hash("not-a-real-password")
	})
	if dummyErr != nil {
		return "", dummyErr
	}
	if dummyHash == "" {
		return "", errors.New("password: empty dummy hash")
	}

	return dummyHash, nil
}

func decodeArgon2id(encoded string) (salt, sum []byte, memory, timeCost uint32, threads uint8, ok bool) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return nil, nil, 0, 0, 0, false
	}

	var parsedThreads uint32
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &parsedThreads); err != nil {
		return nil, nil, 0, 0, 0, false
	}
	if memory == 0 || memory > maxArgonMemory || timeCost == 0 || timeCost > maxArgonTime || parsedThreads == 0 || parsedThreads > uint32(maxArgonThreads) {
		return nil, nil, 0, 0, 0, false
	}

	var err error
	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) == 0 {
		return nil, nil, 0, 0, 0, false
	}
	sum, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(sum) != argonKeyLen {
		return nil, nil, 0, 0, 0, false
	}

	return salt, sum, memory, timeCost, uint8(parsedThreads), true
}
