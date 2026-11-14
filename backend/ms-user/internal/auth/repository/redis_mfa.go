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

const mfaLoginPrefix = "mfa:login:"

// MFAStore manages ephemeral MFA login challenges in Redis.
type MFAStore interface {
	PutLoginMFA(ctx context.Context, challengeID string, rec models.LoginMFAPending, ttl time.Duration) error
	GetLoginMFA(ctx context.Context, challengeID string) (models.LoginMFAPending, error)
	UpdateLoginMFA(ctx context.Context, challengeID string, rec models.LoginMFAPending, ttl time.Duration) error
	DeleteLoginMFA(ctx context.Context, challengeID string) error
}

type redisMFAStore struct {
	rdb *redis.Client
}

// NewMFAStore stores login MFA challenges in Redis.
func NewMFAStore(rdb *redis.Client) MFAStore {
	return redisMFAStore{rdb: rdb}
}

func (s redisMFAStore) PutLoginMFA(ctx context.Context, challengeID string, rec models.LoginMFAPending, ttl time.Duration) error {
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, mfaLoginPrefix+challengeID, raw, ttl).Err()
}

func (s redisMFAStore) GetLoginMFA(ctx context.Context, challengeID string) (models.LoginMFAPending, error) {
	raw, err := s.rdb.Get(ctx, mfaLoginPrefix+challengeID).Bytes()
	if errors.Is(err, redis.Nil) {
		return models.LoginMFAPending{}, constants.ErrNotFound
	}
	if err != nil {
		return models.LoginMFAPending{}, err
	}
	var rec models.LoginMFAPending
	if err := json.Unmarshal(raw, &rec); err != nil {
		return models.LoginMFAPending{}, err
	}
	return rec, nil
}

func (s redisMFAStore) UpdateLoginMFA(ctx context.Context, challengeID string, rec models.LoginMFAPending, ttl time.Duration) error {
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	key := mfaLoginPrefix + challengeID
	rem, err := s.rdb.TTL(ctx, key).Result()
	if err != nil {
		return err
	}
	if rem > 0 {
		ttl = rem
	}
	return s.rdb.Set(ctx, key, raw, ttl).Err()
}

func (s redisMFAStore) DeleteLoginMFA(ctx context.Context, challengeID string) error {
	return s.rdb.Del(ctx, mfaLoginPrefix+challengeID).Err()
}
