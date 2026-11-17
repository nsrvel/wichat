package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-core/internal/models"
)

const (
	roomMetaPrefix         = "voice:room:meta:"
	roomOccupantsPrefix    = "voice:room:occupants:"
	groupChannelsPrefix    = "voice:group:channels:"
	userActiveRoomPrefix   = "voice:user:active_room:"
	tokenRateLimitPrefix   = "voice:ratelimit:token:"
	webhookDedupPrefix     = "voice:webhook:dedup:"
	occupiedChannelsSetKey = "voice:occupied:channels"
	roomStateTTL           = 12 * time.Hour
	userActiveRoomTTL      = 2 * time.Hour
	webhookDedupTTL        = 24 * time.Hour
)

// RedisOccupancy manages live voice room occupancy, rate limits, and webhook dedup in Redis.
type RedisOccupancy interface {
	SetRoomMeta(ctx context.Context, channelUUID string, meta models.RoomMeta) error
	GetRoomMeta(ctx context.Context, channelUUID string) (models.RoomMeta, error)
	UpsertOccupant(ctx context.Context, channelUUID, groupPublicID string, occ models.VoiceOccupant) error
	RemoveOccupant(ctx context.Context, channelUUID, groupPublicID, userUUID string) error
	ListOccupants(ctx context.Context, channelUUID string) ([]models.VoiceOccupant, error)
	CountOccupants(ctx context.Context, channelUUID string) (int, error)
	CountVideoPublishers(ctx context.Context, channelUUID string) (int, error)
	ListGroupChannelUUIDs(ctx context.Context, groupPublicID string) ([]string, error)
	ListAllOccupiedChannelUUIDs(ctx context.Context) ([]string, error)
	ClearRoom(ctx context.Context, channelUUID, groupPublicID string) error
	TouchOccupantActivity(ctx context.Context, channelUUID, userUUID string, nowUnix int64) error
	MarkOccupantIdleWarned(ctx context.Context, channelUUID, userUUID string, warned bool) error
	AllowTokenRequest(ctx context.Context, userUUID string, maxPerMin int64) (bool, error)
	MarkWebhookProcessed(ctx context.Context, eventID string) (bool, error)
}

type redisOccupancy struct {
	rdb *redis.Client
}

// NewRedisOccupancy constructs a Redis-backed voice occupancy store.
func NewRedisOccupancy(rdb *redis.Client) RedisOccupancy {
	return redisOccupancy{rdb: rdb}
}

func (r redisOccupancy) SetRoomMeta(ctx context.Context, channelUUID string, meta models.RoomMeta) error {
	key := roomMetaPrefix + channelUUID
	isRec := "0"
	if meta.IsRecording {
		isRec = "1"
	}
	pipe := r.rdb.TxPipeline()
	pipe.HSet(ctx, key, map[string]any{
		"session_id":            meta.SessionID,
		"group_public_id":       meta.GroupPublicID,
		"channel_public_id":     meta.ChannelPublicID,
		"room_name":             meta.RoomName,
		"started_at":            meta.StartedAt,
		"is_recording":          isRec,
		"active_recording_id":   meta.ActiveRecordingID,
		"video_publisher_count": meta.VideoPublisherCount,
	})
	pipe.Expire(ctx, key, roomStateTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (r redisOccupancy) GetRoomMeta(ctx context.Context, channelUUID string) (models.RoomMeta, error) {
	vals, err := r.rdb.HGetAll(ctx, roomMetaPrefix+channelUUID).Result()
	if err != nil {
		return models.RoomMeta{}, err
	}
	if len(vals) == 0 {
		return models.RoomMeta{}, ErrNotFound
	}
	vidCount, _ := strconv.Atoi(vals["video_publisher_count"])
	return models.RoomMeta{
		SessionID:           vals["session_id"],
		GroupPublicID:       vals["group_public_id"],
		ChannelPublicID:     vals["channel_public_id"],
		RoomName:            vals["room_name"],
		StartedAt:           vals["started_at"],
		IsRecording:         vals["is_recording"] == "1",
		ActiveRecordingID:   vals["active_recording_id"],
		VideoPublisherCount: vidCount,
	}, nil
}

func (r redisOccupancy) UpsertOccupant(ctx context.Context, channelUUID, groupPublicID string, occ models.VoiceOccupant) error {
	raw, err := json.Marshal(occ)
	if err != nil {
		return err
	}
	occKey := roomOccupantsPrefix + channelUUID
	pipe := r.rdb.TxPipeline()
	pipe.HSet(ctx, occKey, occ.UserID, string(raw))
	pipe.Expire(ctx, occKey, roomStateTTL)
	pipe.SAdd(ctx, occupiedChannelsSetKey, channelUUID)
	if groupPublicID != "" {
		gKey := groupChannelsPrefix + groupPublicID
		pipe.SAdd(ctx, gKey, channelUUID)
		pipe.Expire(ctx, gKey, roomStateTTL)
	}
	pipe.Set(ctx, userActiveRoomPrefix+occ.UserID, channelUUID, userActiveRoomTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func (r redisOccupancy) RemoveOccupant(ctx context.Context, channelUUID, groupPublicID, userUUID string) error {
	occKey := roomOccupantsPrefix + channelUUID
	pipe := r.rdb.TxPipeline()
	pipe.HDel(ctx, occKey, userUUID)
	pipe.Del(ctx, userActiveRoomPrefix+userUUID)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	remaining, err := r.rdb.HLen(ctx, occKey).Result()
	if err == nil && remaining == 0 {
		clean := r.rdb.TxPipeline()
		clean.SRem(ctx, occupiedChannelsSetKey, channelUUID)
		if groupPublicID != "" {
			clean.SRem(ctx, groupChannelsPrefix+groupPublicID, channelUUID)
		}
		_, _ = clean.Exec(ctx)
	}
	return nil
}

func (r redisOccupancy) ListOccupants(ctx context.Context, channelUUID string) ([]models.VoiceOccupant, error) {
	vals, err := r.rdb.HGetAll(ctx, roomOccupantsPrefix+channelUUID).Result()
	if err != nil {
		return nil, err
	}
	out := make([]models.VoiceOccupant, 0, len(vals))
	for _, raw := range vals {
		var occ models.VoiceOccupant
		if err := json.Unmarshal([]byte(raw), &occ); err == nil {
			out = append(out, occ)
		}
	}
	return out, nil
}

func (r redisOccupancy) CountOccupants(ctx context.Context, channelUUID string) (int, error) {
	n, err := r.rdb.HLen(ctx, roomOccupantsPrefix+channelUUID).Result()
	return int(n), err
}

func (r redisOccupancy) CountVideoPublishers(ctx context.Context, channelUUID string) (int, error) {
	occs, err := r.ListOccupants(ctx, channelUUID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, o := range occs {
		if o.IsCameraOn || o.IsScreenSharing {
			count++
		}
	}
	return count, nil
}

func (r redisOccupancy) ListGroupChannelUUIDs(ctx context.Context, groupPublicID string) ([]string, error) {
	return r.rdb.SMembers(ctx, groupChannelsPrefix+groupPublicID).Result()
}

func (r redisOccupancy) ListAllOccupiedChannelUUIDs(ctx context.Context) ([]string, error) {
	return r.rdb.SMembers(ctx, occupiedChannelsSetKey).Result()
}

func (r redisOccupancy) ClearRoom(ctx context.Context, channelUUID, groupPublicID string) error {
	occs, _ := r.ListOccupants(ctx, channelUUID)
	pipe := r.rdb.TxPipeline()
	pipe.Del(ctx, roomMetaPrefix+channelUUID)
	pipe.Del(ctx, roomOccupantsPrefix+channelUUID)
	pipe.SRem(ctx, occupiedChannelsSetKey, channelUUID)
	if groupPublicID != "" {
		pipe.SRem(ctx, groupChannelsPrefix+groupPublicID, channelUUID)
	}
	for _, o := range occs {
		pipe.Del(ctx, userActiveRoomPrefix+o.UserID)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r redisOccupancy) TouchOccupantActivity(ctx context.Context, channelUUID, userUUID string, nowUnix int64) error {
	raw, err := r.rdb.HGet(ctx, roomOccupantsPrefix+channelUUID, userUUID).Result()
	if errors.Is(err, redis.Nil) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	var occ models.VoiceOccupant
	if err := json.Unmarshal([]byte(raw), &occ); err != nil {
		return err
	}
	occ.LastActiveAt = nowUnix
	occ.IdleWarned = false
	updated, err := json.Marshal(occ)
	if err != nil {
		return err
	}
	return r.rdb.HSet(ctx, roomOccupantsPrefix+channelUUID, userUUID, string(updated)).Err()
}

func (r redisOccupancy) MarkOccupantIdleWarned(ctx context.Context, channelUUID, userUUID string, warned bool) error {
	raw, err := r.rdb.HGet(ctx, roomOccupantsPrefix+channelUUID, userUUID).Result()
	if errors.Is(err, redis.Nil) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	var occ models.VoiceOccupant
	if err := json.Unmarshal([]byte(raw), &occ); err != nil {
		return err
	}
	occ.IdleWarned = warned
	updated, err := json.Marshal(occ)
	if err != nil {
		return err
	}
	return r.rdb.HSet(ctx, roomOccupantsPrefix+channelUUID, userUUID, string(updated)).Err()
}

func (r redisOccupancy) AllowTokenRequest(ctx context.Context, userUUID string, maxPerMin int64) (bool, error) {
	key := tokenRateLimitPrefix + userUUID
	count, err := r.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		_ = r.rdb.Expire(ctx, key, time.Minute).Err()
	}
	return count <= maxPerMin, nil
}

func (r redisOccupancy) MarkWebhookProcessed(ctx context.Context, eventID string) (bool, error) {
	if eventID == "" {
		return true, nil
	}
	return r.rdb.SetNX(ctx, webhookDedupPrefix+eventID, "1", webhookDedupTTL).Result()
}
