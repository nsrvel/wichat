package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

// EmailChangeStore holds ephemeral email-change OTP challenges and verification tokens in Redis.
type EmailChangeStore interface {
	PutChallenge(ctx context.Context, challengeID string, rec models.EmailChangePending, ttl time.Duration) error
	GetChallenge(ctx context.Context, challengeID string) (models.EmailChangePending, error)
	VerifyChallenge(ctx context.Context, challengeID, code string, secret []byte, maxAttempts int) (models.EmailChangePending, error)
	DeleteChallenge(ctx context.Context, challengeID string) error
	SaveChangeToken(ctx context.Context, token, userID string, ttl time.Duration) error
	ConsumeChangeToken(ctx context.Context, token string) (string, error)
}

type redisEmailChangeStore struct {
	rdb *redis.Client
}

// NewRedisEmailChangeStore stores email-change OTP challenges and tokens in Redis.
func NewRedisEmailChangeStore(rdb *redis.Client) EmailChangeStore {
	return redisEmailChangeStore{rdb: rdb}
}

func emailChangeChallengeKey(challengeID string) string {
	return authRedis.KeyPrefix + "email_change:challenge:" + challengeID
}

func emailChangeTokenKey(token string) string {
	return authRedis.KeyPrefix + "email_change:token:" + token
}

func (s redisEmailChangeStore) PutChallenge(ctx context.Context, challengeID string, rec models.EmailChangePending, ttl time.Duration) error {
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, emailChangeChallengeKey(challengeID), payload, ttl).Err()
}

func (s redisEmailChangeStore) GetChallenge(ctx context.Context, challengeID string) (models.EmailChangePending, error) {
	raw, err := s.rdb.Get(ctx, emailChangeChallengeKey(challengeID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return models.EmailChangePending{}, constants.ErrNotFound
	}
	if err != nil {
		return models.EmailChangePending{}, err
	}
	var rec models.EmailChangePending
	if err := json.Unmarshal(raw, &rec); err != nil {
		return models.EmailChangePending{}, err
	}

	return rec, nil
}

func (s redisEmailChangeStore) VerifyChallenge(ctx context.Context, challengeID, code string, secret []byte, maxAttempts int) (models.EmailChangePending, error) {
	key := emailChangeChallengeKey(challengeID)
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return models.EmailChangePending{}, constants.ErrNotFound
	}
	if err != nil {
		return models.EmailChangePending{}, err
	}

	var rec models.EmailChangePending
	if err := json.Unmarshal(raw, &rec); err != nil {
		return models.EmailChangePending{}, err
	}

	if rec.CodeHMAC != authRedis.HMACCode(secret, authRedis.NormalizeOTP(code)) {
		rec.Attempts++
		if rec.Attempts >= maxAttempts {
			_ = s.rdb.Del(ctx, key).Err()
			return models.EmailChangePending{}, constants.ErrNotFound
		}
		payload, err := json.Marshal(rec)
		if err != nil {
			return models.EmailChangePending{}, err
		}
		ttl, err := s.rdb.TTL(ctx, key).Result()
		if err != nil {
			return models.EmailChangePending{}, err
		}
		if ttl <= 0 {
			ttl = 10 * time.Minute
		}
		if err := s.rdb.Set(ctx, key, payload, ttl).Err(); err != nil {
			return models.EmailChangePending{}, err
		}

		return models.EmailChangePending{}, constants.ErrNotFound
	}

	return rec, nil
}

func (s redisEmailChangeStore) DeleteChallenge(ctx context.Context, challengeID string) error {
	return s.rdb.Del(ctx, emailChangeChallengeKey(challengeID)).Err()
}

func (s redisEmailChangeStore) SaveChangeToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	payload, err := json.Marshal(models.EmailChangeTokenRecord{UserID: userID})
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, emailChangeTokenKey(token), payload, ttl).Err()
}

func (s redisEmailChangeStore) ConsumeChangeToken(ctx context.Context, token string) (string, error) {
	raw, err := s.rdb.GetDel(ctx, emailChangeTokenKey(token)).Bytes()
	if errors.Is(err, redis.Nil) {
		return "", constants.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	var rec models.EmailChangeTokenRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return "", err
	}

	return rec.UserID, nil
}
