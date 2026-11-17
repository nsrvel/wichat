package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
)

const (
	connKeyPrefix          = "presence:conn:"
	connIDsKeyPrefix       = "presence:conn_ids:"
	manualKeyPrefix        = "presence:manual:"
	effectiveKeyPrefix     = "presence:effective:"
	voiceActivityKeyPrefix = "presence:voice_activity:"
)

// RedisRepository stores ephemeral connection heartbeats, manual overrides, and effective presence in Redis.
type RedisRepository interface {
	SetConnHeartbeat(ctx context.Context, userID, connID, activity string, ttl time.Duration) error
	DeleteConn(ctx context.Context, userID, connID string) error
	ListActiveConnActivities(ctx context.Context, userID string) ([]string, error)
	SetManualStatus(ctx context.Context, userID, status string, ttl time.Duration) error
	GetManualStatus(ctx context.Context, userID string) (models.ManualPresenceState, error)
	SetEffectiveStatus(ctx context.Context, userID, effectiveStatus string, ttl time.Duration) error
	GetEffectiveStatus(ctx context.Context, userID string) (string, error)
	BatchGetEffectivePresence(ctx context.Context, userIDs []string) (map[string]models.UserPresenceSnapshot, error)
	SetVoiceActivityStatus(ctx context.Context, userID, activityStatus string, ttl time.Duration) error
	GetVoiceActivityStatus(ctx context.Context, userID string) (string, error)
}

type redisPresenceRepo struct {
	rdb *redis.Client
}

// NewRedisRepository constructs a Redis-backed presence repository.
func NewRedisRepository(rdb *redis.Client) RedisRepository {
	return redisPresenceRepo{rdb: rdb}
}

func connKey(userID, connID string) string {
	return connKeyPrefix + userID + ":" + connID
}

func connIDsKey(userID string) string {
	return connIDsKeyPrefix + userID
}

func manualKey(userID string) string {
	return manualKeyPrefix + userID
}

func effectiveKey(userID string) string {
	return effectiveKeyPrefix + userID
}

func (r redisPresenceRepo) SetConnHeartbeat(ctx context.Context, userID, connID, activity string, ttl time.Duration) error {
	pipe := r.rdb.TxPipeline()
	pipe.Set(ctx, connKey(userID, connID), activity, ttl)
	pipe.SAdd(ctx, connIDsKey(userID), connID)
	pipe.Expire(ctx, connIDsKey(userID), ttl*2)
	_, err := pipe.Exec(ctx)
	return err
}

func (r redisPresenceRepo) DeleteConn(ctx context.Context, userID, connID string) error {
	pipe := r.rdb.TxPipeline()
	pipe.Del(ctx, connKey(userID, connID))
	pipe.SRem(ctx, connIDsKey(userID), connID)
	_, err := pipe.Exec(ctx)
	return err
}

func (r redisPresenceRepo) ListActiveConnActivities(ctx context.Context, userID string) ([]string, error) {
	ids, err := r.rdb.SMembers(ctx, connIDsKey(userID)).Result()
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = connKey(userID, id)
	}

	vals, err := r.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	activities := make([]string, 0, len(vals))
	var expired []any
	for i, v := range vals {
		if v == nil {
			expired = append(expired, ids[i])
			continue
		}
		if s, ok := v.(string); ok && s != "" {
			activities = append(activities, s)
		}
	}
	if len(expired) > 0 {
		_ = r.rdb.SRem(ctx, connIDsKey(userID), expired...).Err()
	}

	return activities, nil
}

func (r redisPresenceRepo) SetManualStatus(ctx context.Context, userID, status string, ttl time.Duration) error {
	if status == "" || status == models.PresenceStatusOnline {
		return r.rdb.Del(ctx, manualKey(userID)).Err()
	}
	return r.rdb.Set(ctx, manualKey(userID), status, ttl).Err()
}

func (r redisPresenceRepo) GetManualStatus(ctx context.Context, userID string) (models.ManualPresenceState, error) {
	pipe := r.rdb.TxPipeline()
	getCmd := pipe.Get(ctx, manualKey(userID))
	ttlCmd := pipe.TTL(ctx, manualKey(userID))
	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return models.ManualPresenceState{}, err
	}

	val, getErr := getCmd.Result()
	if errors.Is(getErr, redis.Nil) || val == "" {
		return models.ManualPresenceState{Status: models.PresenceStatusOnline}, nil
	}
	if getErr != nil {
		return models.ManualPresenceState{}, getErr
	}

	var expiresAt *time.Time
	if ttl := ttlCmd.Val(); ttl > 0 {
		t := time.Now().UTC().Add(ttl)
		expiresAt = &t
	}
	return models.ManualPresenceState{Status: val, ExpiresAt: expiresAt}, nil
}

func (r redisPresenceRepo) SetEffectiveStatus(ctx context.Context, userID, effectiveStatus string, ttl time.Duration) error {
	if effectiveStatus == "" || effectiveStatus == models.PresenceStatusOffline {
		return r.rdb.Del(ctx, effectiveKey(userID)).Err()
	}
	return r.rdb.Set(ctx, effectiveKey(userID), effectiveStatus, ttl).Err()
}

func (r redisPresenceRepo) GetEffectiveStatus(ctx context.Context, userID string) (string, error) {
	val, err := r.rdb.Get(ctx, effectiveKey(userID)).Result()
	if errors.Is(err, redis.Nil) || val == "" {
		return models.PresenceStatusOffline, nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r redisPresenceRepo) BatchGetEffectivePresence(ctx context.Context, userIDs []string) (map[string]models.UserPresenceSnapshot, error) {
	out := make(map[string]models.UserPresenceSnapshot, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}

	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = effectiveKey(id)
	}

	vals, err := r.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	for i, id := range userIDs {
		status := models.PresenceStatusOffline
		if i < len(vals) {
			if s, ok := vals[i].(string); ok && s != "" {
				status = s
			}
		}
		out[id] = models.UserPresenceSnapshot{
			IsOnline:       status != models.PresenceStatusOffline,
			PresenceStatus: status,
		}
	}

	return out, nil
}

func (r redisPresenceRepo) SetVoiceActivityStatus(ctx context.Context, userID, activityStatus string, ttl time.Duration) error {
	key := voiceActivityKeyPrefix + userID
	if activityStatus == "" {
		return r.rdb.Del(ctx, key).Err()
	}
	return r.rdb.Set(ctx, key, activityStatus, ttl).Err()
}

func (r redisPresenceRepo) GetVoiceActivityStatus(ctx context.Context, userID string) (string, error) {
	val, err := r.rdb.Get(ctx, voiceActivityKeyPrefix+userID).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}
