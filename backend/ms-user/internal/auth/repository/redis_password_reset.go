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

// PasswordResetStore holds ephemeral password-reset state in Redis.
type PasswordResetStore interface {
	EnsureCanSendResetCode(ctx context.Context, email string) error
	RecordResetCodeSent(ctx context.Context, email string) (time.Duration, error)
	RecordWrongResetCode(ctx context.Context, email string) error
	ClearResendPenalty(ctx context.Context, email string) error
	SaveOTP(ctx context.Context, email, userID, code string, secret []byte, ttl time.Duration) error
	VerifyOTP(ctx context.Context, email, code string, secret []byte, maxAttempts int) (userID string, err error)
	DeleteOTP(ctx context.Context, email string) error
	SaveResetToken(ctx context.Context, token, userID, email string, ttl time.Duration) error
	ConsumeResetToken(ctx context.Context, token string) (userID, email string, err error)
}

type redisPasswordResetStore struct {
	rdb *redis.Client
}

// NewRedisPasswordResetStore stores reset OTP and tokens in Redis.
func NewRedisPasswordResetStore(rdb *redis.Client) PasswordResetStore {
	return redisPasswordResetStore{rdb: rdb}
}

func (s redisPasswordResetStore) EnsureCanSendResetCode(ctx context.Context, email string) error {
	ttl, err := s.rdb.TTL(ctx, pwResetCooldownKey(email)).Result()
	if err != nil {
		return err
	}
	if ttl > 0 {
		return &models.ResetRateLimitError{RetryAfter: ttl}
	}

	return nil
}

// RecordResetCodeSent increments resend strikes and starts the next wait window.
func (s redisPasswordResetStore) RecordResetCodeSent(ctx context.Context, email string) (time.Duration, error) {
	strikes, err := s.bumpStrikes(ctx, email)
	if err != nil {
		return 0, err
	}

	return s.applyCooldown(ctx, email, strikes)
}

// RecordWrongResetCode lengthens the resend wait after a bad OTP attempt.
func (s redisPasswordResetStore) RecordWrongResetCode(ctx context.Context, email string) error {
	strikes, err := s.bumpStrikes(ctx, email)
	if err != nil {
		return err
	}

	return s.extendCooldown(ctx, email, authRedis.ResendCooldownForStrikes(strikes))
}

func (s redisPasswordResetStore) ClearResendPenalty(ctx context.Context, email string) error {
	return s.rdb.Del(ctx, pwResetCooldownKey(email), pwResetStrikesKey(email)).Err()
}

func (s redisPasswordResetStore) SaveOTP(ctx context.Context, email, userID, code string, secret []byte, ttl time.Duration) error {
	rec := models.PasswordResetOTP{
		UserID:   userID,
		CodeHMAC: authRedis.HMACCode(secret, code),
		Attempts: 0,
	}
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, pwResetOTPKey(email), payload, ttl).Err()
}

func (s redisPasswordResetStore) VerifyOTP(ctx context.Context, email, code string, secret []byte, maxAttempts int) (string, error) {
	key := pwResetOTPKey(email)
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return "", constants.ErrNotFound
	}
	if err != nil {
		return "", err
	}

	var rec models.PasswordResetOTP
	if err := json.Unmarshal(raw, &rec); err != nil {
		return "", err
	}

	if rec.CodeHMAC != authRedis.HMACCode(secret, authRedis.NormalizeOTP(code)) {
		rec.Attempts++
		if rec.Attempts >= maxAttempts {
			_ = s.rdb.Del(ctx, key).Err()
			return "", constants.ErrNotFound
		}
		payload, err := json.Marshal(rec)
		if err != nil {
			return "", err
		}
		ttl, err := s.rdb.TTL(ctx, key).Result()
		if err != nil {
			return "", err
		}
		if ttl <= 0 {
			ttl = 10 * time.Minute
		}
		if err := s.rdb.Set(ctx, key, payload, ttl).Err(); err != nil {
			return "", err
		}

		return "", constants.ErrNotFound
	}

	return rec.UserID, nil
}

func (s redisPasswordResetStore) DeleteOTP(ctx context.Context, email string) error {
	return s.rdb.Del(ctx, pwResetOTPKey(email)).Err()
}

func (s redisPasswordResetStore) SaveResetToken(ctx context.Context, token, userID, email string, ttl time.Duration) error {
	rec := models.PasswordResetTokenRecord{UserID: userID, Email: email}
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, pwResetTokenKey(token), payload, ttl).Err()
}

func (s redisPasswordResetStore) ConsumeResetToken(ctx context.Context, token string) (string, string, error) {
	key := pwResetTokenKey(token)
	raw, err := s.rdb.GetDel(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return "", "", constants.ErrNotFound
	}
	if err != nil {
		return "", "", err
	}
	var rec models.PasswordResetTokenRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return "", "", err
	}

	return rec.UserID, rec.Email, nil
}

func pwResetCooldownKey(email string) string {
	return "auth:pwreset:cooldown:" + authRedis.EmailKey(email)
}

func pwResetStrikesKey(email string) string {
	return "auth:pwreset:strikes:" + authRedis.EmailKey(email)
}

func pwResetOTPKey(email string) string {
	return "auth:pwreset:otp:" + authRedis.EmailKey(email)
}

func pwResetTokenKey(token string) string {
	return "auth:pwreset:token:" + token
}

func (s redisPasswordResetStore) bumpStrikes(ctx context.Context, email string) (int, error) {
	key := pwResetStrikesKey(email)
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

func (s redisPasswordResetStore) applyCooldown(ctx context.Context, email string, strikes int) (time.Duration, error) {
	d := authRedis.ResendCooldownForStrikes(strikes)
	if d <= 0 {
		return 0, nil
	}
	if err := s.rdb.Set(ctx, pwResetCooldownKey(email), "1", d).Err(); err != nil {
		return 0, err
	}

	return d, nil
}

func (s redisPasswordResetStore) extendCooldown(ctx context.Context, email string, minWait time.Duration) error {
	key := pwResetCooldownKey(email)
	ttl, err := s.rdb.TTL(ctx, key).Result()
	if err != nil {
		return err
	}
	if ttl >= minWait {
		return nil
	}

	return s.rdb.Set(ctx, key, "1", minWait).Err()
}
