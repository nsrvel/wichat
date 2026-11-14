package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

const (
	totpSetupPrefix        = "mfa:totp_setup:"
	webAuthnRegisterPrefix = "mfa:webauthn_reg:"
)

// MFAEnrollmentStore manages ephemeral TOTP setup and WebAuthn registration challenges in Redis.
type MFAEnrollmentStore interface {
	PutTOTPSetup(ctx context.Context, token string, rec models.TOTPSetupPending, ttl time.Duration) error
	TakeTOTPSetup(ctx context.Context, token string) (models.TOTPSetupPending, error)
	PutWebAuthnRegister(ctx context.Context, challengeID string, rec models.WebAuthnRegisterPending, ttl time.Duration) error
	TakeWebAuthnRegister(ctx context.Context, challengeID string) (models.WebAuthnRegisterPending, error)
}

type redisMFAEnrollmentStore struct {
	rdb *redis.Client
}

// NewMFAEnrollmentStore stores ephemeral MFA enrollment challenges in Redis.
func NewMFAEnrollmentStore(rdb *redis.Client) MFAEnrollmentStore {
	return redisMFAEnrollmentStore{rdb: rdb}
}

func (s redisMFAEnrollmentStore) PutTOTPSetup(ctx context.Context, token string, rec models.TOTPSetupPending, ttl time.Duration) error {
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, totpSetupPrefix+token, raw, ttl).Err()
}

func (s redisMFAEnrollmentStore) TakeTOTPSetup(ctx context.Context, token string) (models.TOTPSetupPending, error) {
	raw, err := s.rdb.GetDel(ctx, totpSetupPrefix+token).Bytes()
	if errors.Is(err, redis.Nil) {
		return models.TOTPSetupPending{}, constants.ErrNotFound
	}
	if err != nil {
		return models.TOTPSetupPending{}, err
	}
	var rec models.TOTPSetupPending
	if err := json.Unmarshal(raw, &rec); err != nil {
		return models.TOTPSetupPending{}, err
	}
	return rec, nil
}

func (s redisMFAEnrollmentStore) PutWebAuthnRegister(ctx context.Context, challengeID string, rec models.WebAuthnRegisterPending, ttl time.Duration) error {
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, webAuthnRegisterPrefix+challengeID, raw, ttl).Err()
}

func (s redisMFAEnrollmentStore) TakeWebAuthnRegister(ctx context.Context, challengeID string) (models.WebAuthnRegisterPending, error) {
	raw, err := s.rdb.GetDel(ctx, webAuthnRegisterPrefix+challengeID).Bytes()
	if errors.Is(err, redis.Nil) {
		return models.WebAuthnRegisterPending{}, constants.ErrNotFound
	}
	if err != nil {
		return models.WebAuthnRegisterPending{}, err
	}
	var rec models.WebAuthnRegisterPending
	if err := json.Unmarshal(raw, &rec); err != nil {
		return models.WebAuthnRegisterPending{}, err
	}
	return rec, nil
}
