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

// EmailAuthStore holds register/login email OTP challenges in Redis.
type EmailAuthStore interface {
	EnsureCanSendEmailCode(ctx context.Context, email string) error
	RecordEmailCodeSent(ctx context.Context, email string) (time.Duration, error)
	RecordWrongEmailCode(ctx context.Context, email string) error
	ClearEmailCodePenalty(ctx context.Context, email string) error

	PutRegisterPending(ctx context.Context, challengeID string, rec models.RegisterPending, ttl time.Duration) error
	GetRegisterPending(ctx context.Context, challengeID string) (models.RegisterPending, error)
	VerifyRegisterCode(ctx context.Context, challengeID, code string, secret []byte, maxAttempts int) (models.RegisterPending, error)
	DeleteRegisterPending(ctx context.Context, challengeID string) error
	UpdateRegisterCode(ctx context.Context, challengeID, codeHMAC string, ttl time.Duration) error

	PutLoginChallenge(ctx context.Context, challengeID string, rec models.LoginPending, ttl time.Duration) error
	GetLoginChallenge(ctx context.Context, challengeID string) (models.LoginPending, error)
	VerifyLoginCode(ctx context.Context, challengeID, code string, secret []byte, maxAttempts int) (models.LoginPending, error)
	DeleteLoginChallenge(ctx context.Context, challengeID string) error
	UpdateLoginCode(ctx context.Context, challengeID, codeHMAC string, ttl time.Duration) error
}

type redisEmailAuthStore struct {
	rdb *redis.Client
}

// NewRedisEmailAuthStore stores email OTP challenges in Redis.
func NewRedisEmailAuthStore(rdb *redis.Client) EmailAuthStore {
	return redisEmailAuthStore{rdb: rdb}
}

func (s redisEmailAuthStore) EnsureCanSendEmailCode(ctx context.Context, email string) error {
	ttl, err := s.rdb.TTL(ctx, emailCodeCooldownKey(email)).Result()
	if err != nil {
		return err
	}
	if ttl > 0 {
		return &models.ResetRateLimitError{RetryAfter: ttl}
	}

	return nil
}

func (s redisEmailAuthStore) RecordEmailCodeSent(ctx context.Context, email string) (time.Duration, error) {
	strikes, err := s.bumpEmailStrikes(ctx, email)
	if err != nil {
		return 0, err
	}
	d := authRedis.ResendCooldownForStrikes(strikes)
	if d <= 0 {
		return 0, nil
	}
	if err := s.rdb.Set(ctx, emailCodeCooldownKey(email), "1", d).Err(); err != nil {
		return 0, err
	}

	return d, nil
}

func (s redisEmailAuthStore) RecordWrongEmailCode(ctx context.Context, email string) error {
	strikes, err := s.bumpEmailStrikes(ctx, email)
	if err != nil {
		return err
	}
	d := authRedis.ResendCooldownForStrikes(strikes)
	if d <= 0 {
		return nil
	}
	key := emailCodeCooldownKey(email)
	ttl, err := s.rdb.TTL(ctx, key).Result()
	if err != nil {
		return err
	}
	if ttl >= d {
		return nil
	}

	return s.rdb.Set(ctx, key, "1", d).Err()
}

func (s redisEmailAuthStore) ClearEmailCodePenalty(ctx context.Context, email string) error {
	return s.rdb.Del(ctx, emailCodeCooldownKey(email), emailCodeStrikesKey(email)).Err()
}

func (s redisEmailAuthStore) PutRegisterPending(ctx context.Context, challengeID string, rec models.RegisterPending, ttl time.Duration) error {
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, registerPendingKey(challengeID), payload, ttl).Err()
}

func (s redisEmailAuthStore) GetRegisterPending(ctx context.Context, challengeID string) (models.RegisterPending, error) {
	raw, err := s.rdb.Get(ctx, registerPendingKey(challengeID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return models.RegisterPending{}, constants.ErrNotFound
	}
	if err != nil {
		return models.RegisterPending{}, err
	}
	var rec models.RegisterPending
	if err := json.Unmarshal(raw, &rec); err != nil {
		return models.RegisterPending{}, err
	}

	return rec, nil
}

func (s redisEmailAuthStore) VerifyRegisterCode(ctx context.Context, challengeID, code string, secret []byte, maxAttempts int) (models.RegisterPending, error) {
	key := registerPendingKey(challengeID)
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return models.RegisterPending{}, constants.ErrNotFound
	}
	if err != nil {
		return models.RegisterPending{}, err
	}

	var rec models.RegisterPending
	if err := json.Unmarshal(raw, &rec); err != nil {
		return models.RegisterPending{}, err
	}

	if rec.CodeHMAC != authRedis.HMACCode(secret, authRedis.NormalizeOTP(code)) {
		rec.Attempts++
		if rec.Attempts >= maxAttempts {
			_ = s.rdb.Del(ctx, key).Err()
			return models.RegisterPending{}, constants.ErrNotFound
		}
		payload, err := json.Marshal(rec)
		if err != nil {
			return models.RegisterPending{}, err
		}
		ttl, err := s.rdb.TTL(ctx, key).Result()
		if err != nil {
			return models.RegisterPending{}, err
		}
		if ttl <= 0 {
			ttl = 10 * time.Minute
		}
		if err := s.rdb.Set(ctx, key, payload, ttl).Err(); err != nil {
			return models.RegisterPending{}, err
		}

		return models.RegisterPending{}, constants.ErrNotFound
	}

	return rec, nil
}

func (s redisEmailAuthStore) DeleteRegisterPending(ctx context.Context, challengeID string) error {
	return s.rdb.Del(ctx, registerPendingKey(challengeID)).Err()
}

func (s redisEmailAuthStore) UpdateRegisterCode(ctx context.Context, challengeID, codeHMAC string, ttl time.Duration) error {
	rec, err := s.GetRegisterPending(ctx, challengeID)
	if err != nil {
		return err
	}
	rec.CodeHMAC = codeHMAC
	rec.Attempts = 0
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, registerPendingKey(challengeID), payload, ttl).Err()
}

func (s redisEmailAuthStore) PutLoginChallenge(ctx context.Context, challengeID string, rec models.LoginPending, ttl time.Duration) error {
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, loginChallengeKey(challengeID), payload, ttl).Err()
}

func (s redisEmailAuthStore) GetLoginChallenge(ctx context.Context, challengeID string) (models.LoginPending, error) {
	raw, err := s.rdb.Get(ctx, loginChallengeKey(challengeID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return models.LoginPending{}, constants.ErrNotFound
	}
	if err != nil {
		return models.LoginPending{}, err
	}
	var rec models.LoginPending
	if err := json.Unmarshal(raw, &rec); err != nil {
		return models.LoginPending{}, err
	}

	return rec, nil
}

func (s redisEmailAuthStore) VerifyLoginCode(ctx context.Context, challengeID, code string, secret []byte, maxAttempts int) (models.LoginPending, error) {
	key := loginChallengeKey(challengeID)
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return models.LoginPending{}, constants.ErrNotFound
	}
	if err != nil {
		return models.LoginPending{}, err
	}

	var rec models.LoginPending
	if err := json.Unmarshal(raw, &rec); err != nil {
		return models.LoginPending{}, err
	}

	if rec.CodeHMAC != authRedis.HMACCode(secret, authRedis.NormalizeOTP(code)) {
		rec.Attempts++
		if rec.Attempts >= maxAttempts {
			_ = s.rdb.Del(ctx, key).Err()
			return models.LoginPending{}, constants.ErrNotFound
		}
		payload, err := json.Marshal(rec)
		if err != nil {
			return models.LoginPending{}, err
		}
		ttl, err := s.rdb.TTL(ctx, key).Result()
		if err != nil {
			return models.LoginPending{}, err
		}
		if ttl <= 0 {
			ttl = 10 * time.Minute
		}
		if err := s.rdb.Set(ctx, key, payload, ttl).Err(); err != nil {
			return models.LoginPending{}, err
		}

		return models.LoginPending{}, constants.ErrNotFound
	}

	return rec, nil
}

func (s redisEmailAuthStore) DeleteLoginChallenge(ctx context.Context, challengeID string) error {
	return s.rdb.Del(ctx, loginChallengeKey(challengeID)).Err()
}

func (s redisEmailAuthStore) UpdateLoginCode(ctx context.Context, challengeID, codeHMAC string, ttl time.Duration) error {
	rec, err := s.GetLoginChallenge(ctx, challengeID)
	if err != nil {
		return err
	}
	rec.CodeHMAC = codeHMAC
	rec.Attempts = 0
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, loginChallengeKey(challengeID), payload, ttl).Err()
}

func emailCodeCooldownKey(email string) string {
	return "auth:emailcode:cooldown:" + authRedis.EmailKey(email)
}

func emailCodeStrikesKey(email string) string {
	return "auth:emailcode:strikes:" + authRedis.EmailKey(email)
}

func registerPendingKey(challengeID string) string {
	return "auth:register:pending:" + challengeID
}

func loginChallengeKey(challengeID string) string {
	return "auth:login:challenge:" + challengeID
}

func (s redisEmailAuthStore) bumpEmailStrikes(ctx context.Context, email string) (int, error) {
	key := emailCodeStrikesKey(email)
	n, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 {
		if err := s.rdb.Expire(ctx, key, authRedis.StrikeKeyTTL).Err(); err != nil {
			return 0, err
		}
	}

	return int(n), nil
}
