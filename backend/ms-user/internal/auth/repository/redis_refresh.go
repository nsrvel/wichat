package repository

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

const (
	redisKeyRefreshPrefix = authRedis.KeyPrefix + "rt:"
	redisKeyUsedPrefix    = authRedis.KeyPrefix + "rt:used:"
	redisKeySessionPrefix = authRedis.KeyPrefix + "sess:"
)

type refreshPayload struct {
	SessionID  string `json:"session_id"`
	UserID     string `json:"user_id"`
	RefreshID  string `json:"refresh_id"`
	UsedAtUnix int64  `json:"used_at_unix,omitempty"`
}

// RefreshStore persists refresh token hashes in Redis with rotation and reuse detection.
type RefreshStore interface {
	Put(ctx context.Context, refresh models.RefreshToken, userID string, expiresAt time.Time) error
	FindByHash(ctx context.Context, hash []byte) (models.RefreshToken, error)
	Rotate(ctx context.Context, currentHash, nextHash []byte, nextRefresh models.RefreshToken, userID string, sessionExpiresAt, now time.Time) error
	RevokeForSession(ctx context.Context, sessionID string) error
}

type redisRefreshStore struct {
	rdb *redis.Client
}

// NewRedisRefreshStore stores refresh secrets in Redis.
func NewRedisRefreshStore(rdb *redis.Client) RefreshStore {
	return &redisRefreshStore{rdb: rdb}
}

func (s *redisRefreshStore) Put(ctx context.Context, refresh models.RefreshToken, userID string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return constants.ErrNotFound
	}

	payload, err := json.Marshal(refreshPayload{
		SessionID: refresh.SessionID,
		UserID:    userID,
		RefreshID: refresh.ID,
	})
	if err != nil {
		return err
	}

	hashKey := redisKeyRefreshPrefix + hex.EncodeToString(refresh.TokenHash)
	sessKey := redisKeySessionPrefix + refresh.SessionID

	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, hashKey, payload, ttl)
	pipe.Set(ctx, sessKey, hex.EncodeToString(refresh.TokenHash), ttl)
	_, err = pipe.Exec(ctx)

	return err
}

func (s *redisRefreshStore) FindByHash(ctx context.Context, hash []byte) (models.RefreshToken, error) {
	hashKey := redisKeyRefreshPrefix + hex.EncodeToString(hash)
	raw, err := s.rdb.Get(ctx, hashKey).Bytes()
	if err == nil {
		var p refreshPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return models.RefreshToken{}, err
		}

		return models.RefreshToken{
			ID:        p.RefreshID,
			SessionID: p.SessionID,
			TokenHash: hash,
		}, nil
	}
	if !errors.Is(err, redis.Nil) {
		return models.RefreshToken{}, err
	}

	usedKey := redisKeyUsedPrefix + hex.EncodeToString(hash)
	if s.rdb.Exists(ctx, usedKey).Val() > 0 {
		var p refreshPayload
		if rawUsed, err := s.rdb.Get(ctx, usedKey).Bytes(); err == nil {
			_ = json.Unmarshal(rawUsed, &p)
		}
		used := time.Now().UTC()
		if p.UsedAtUnix > 0 {
			used = time.Unix(p.UsedAtUnix, 0).UTC()
		}

		return models.RefreshToken{
			ID:        p.RefreshID,
			SessionID: p.SessionID,
			TokenHash: hash,
			UsedAt:    &used,
		}, nil
	}

	return models.RefreshToken{}, constants.ErrNotFound
}

func (s *redisRefreshStore) Rotate(ctx context.Context, currentHash, nextHash []byte, nextRefresh models.RefreshToken, userID string, sessionExpiresAt, now time.Time) error {
	rec, err := s.FindByHash(ctx, currentHash)
	if errors.Is(err, constants.ErrNotFound) {
		return constants.ErrNotFound
	}
	if err != nil {
		return err
	}
	if rec.UsedAt != nil {
		return constants.ErrAlreadyUsed
	}

	ttl := time.Until(sessionExpiresAt)
	if ttl <= 0 {
		return constants.ErrNotFound
	}

	currentKey := redisKeyRefreshPrefix + hex.EncodeToString(currentHash)
	usedKey := redisKeyUsedPrefix + hex.EncodeToString(currentHash)
	usedPayload, _ := json.Marshal(refreshPayload{
		SessionID:  rec.SessionID,
		UserID:     userID,
		RefreshID:  rec.ID,
		UsedAtUnix: now.Unix(),
	})

	nextKey := redisKeyRefreshPrefix + hex.EncodeToString(nextHash)
	nextPayload, err := json.Marshal(refreshPayload{
		SessionID: nextRefresh.SessionID,
		UserID:    userID,
		RefreshID: nextRefresh.ID,
	})
	if err != nil {
		return err
	}

	sessKey := redisKeySessionPrefix + rec.SessionID

	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, currentKey)
	pipe.Set(ctx, usedKey, usedPayload, ttl)
	pipe.Set(ctx, nextKey, nextPayload, ttl)
	pipe.Set(ctx, sessKey, hex.EncodeToString(nextHash), ttl)
	_, err = pipe.Exec(ctx)

	return err
}

func (s *redisRefreshStore) RevokeForSession(ctx context.Context, sessionID string) error {
	sessKey := redisKeySessionPrefix + sessionID
	hashHex, err := s.rdb.Get(ctx, sessKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}

	hashKey := redisKeyRefreshPrefix + hashHex
	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, sessKey)
	pipe.Del(ctx, hashKey)
	_, err = pipe.Exec(ctx)

	return err
}
