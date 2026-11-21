package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
)

// DevOTPStore stores plaintext OTPs for local e2e (development wiring only).
type DevOTPStore interface {
	Put(ctx context.Context, challengeID, code string, ttl time.Duration) error
	Get(ctx context.Context, challengeID string) (string, error)
}

type redisDevOTPStore struct {
	rdb *redis.Client
}

// NewRedisDevOTPStore writes dev-only e2e OTP shadows.
func NewRedisDevOTPStore(rdb *redis.Client) DevOTPStore {
	return redisDevOTPStore{rdb: rdb}
}

func (s redisDevOTPStore) Put(ctx context.Context, challengeID, code string, ttl time.Duration) error {
	if challengeID == "" || code == "" {
		return nil
	}
	return s.rdb.Set(ctx, authRedis.DevE2EOTPKey(challengeID), authRedis.NormalizeOTP(code), ttl).Err()
}

func (s redisDevOTPStore) Get(ctx context.Context, challengeID string) (string, error) {
	return s.rdb.Get(ctx, authRedis.DevE2EOTPKey(challengeID)).Result()
}
