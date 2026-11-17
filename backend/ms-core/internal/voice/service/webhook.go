package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/livekit/protocol/livekit"
	"github.com/wichat/wichat/backend/ms-core/internal/models"
	"github.com/wichat/wichat/backend/wi-shared/chat-events"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func parseRoomChannelUUID(roomName string) string {
	if strings.HasPrefix(roomName, "channel:") {
		return strings.TrimPrefix(roomName, "channel:")
	}
	if strings.HasPrefix(roomName, "dm:") {
		return strings.TrimPrefix(roomName, "dm:")
	}
	return ""
}

func (s service) HandleWebhook(ctx context.Context, authHeader string, rawBody []byte) error {
	ev, err := s.lk.VerifyWebhook(authHeader, rawBody)
	if err != nil {
		return exception.Unauthenticated(exception.MsgVoiceWebhookInvalidSig, nil)
	}

	if s.redis != nil && ev.GetId() != "" {
		fresh, dErr := s.redis.MarkWebhookProcessed(ctx, ev.GetId())
		if dErr != nil {
			return dErr
		}
		if !fresh {
			return nil
		}
	}

	now := s.now()
	roomName := ev.GetRoom().GetName()
	channelUUID := parseRoomChannelUUID(roomName)

	switch ev.GetEvent() {
	case "room_started":
		if channelUUID == "" {
			return nil
		}
		return s.onRoomStarted(ctx, channelUUID, roomName, now)
	case "participant_joined":
		if channelUUID == "" || ev.GetParticipant() == nil {
			return nil
		}
		return s.onParticipantJoined(ctx, channelUUID, roomName, ev.GetParticipant(), now)
	case "participant_left":
		if channelUUID == "" || ev.GetParticipant() == nil {
			return nil
		}
		return s.onParticipantLeft(ctx, channelUUID, ev.GetParticipant().GetIdentity(), now)
	case "track_published", "track_unpublished":
		if channelUUID == "" || ev.GetParticipant() == nil || ev.GetTrack() == nil {
			return nil
		}
		return s.onTrackChanged(ctx, channelUUID, roomName, ev.GetParticipant().GetIdentity(), ev.GetTrack(), ev.GetEvent() == "track_published", now)
	case "room_finished":
		if channelUUID == "" {
			return nil
		}
		return s.onRoomFinished(ctx, channelUUID, now)
	case "egress_started", "egress_updated", "egress_ended":
		if ev.GetEgressInfo() == nil {
			return nil
		}
		return s.onEgressEvent(ctx, ev.GetEgressInfo(), now)
	default:
		return nil
	}
}

func (s service) onRoomStarted(ctx context.Context, channelUUID, roomName string, now time.Time) error {
	var meta models.RoomMeta
	if s.redis != nil {
		if existing, err := s.redis.GetRoomMeta(ctx, channelUUID); err == nil {
			meta = existing
		}
	}

	sess, err := s.pg.UpsertActiveSession(ctx, models.CallSession{
		ID:               uuid.NewString(),
		PublicID:         newPublicID24(),
		ChannelID:        channelUUID,
		RoomName:         roomName,
		StartedAt:        now,
		PeakParticipants: 1,
	})
	if err != nil {
		return err
	}

	if s.redis != nil {
		meta.SessionID = sess.ID
		meta.RoomName = roomName
		meta.StartedAt = now.Format(time.RFC3339Nano)
		_ = s.redis.SetRoomMeta(ctx, channelUUID, meta)
	}
	return nil
}

func (s service) onParticipantJoined(ctx context.Context, channelUUID, roomName string, p *livekit.ParticipantInfo, now time.Time) error {
	sess, err := s.pg.UpsertActiveSession(ctx, models.CallSession{
		ID:               uuid.NewString(),
		PublicID:         newPublicID24(),
		ChannelID:        channelUUID,
		RoomName:         roomName,
		StartedAt:        now,
		PeakParticipants: 1,
	})
	if err != nil {
		return err
	}

	var parsedMeta struct {
		DisplayName string `json:"display_name"`
		Username    string `json:"username"`
		AvatarURL   string `json:"avatar_url"`
	}
	if p.GetMetadata() != "" {
		_ = json.Unmarshal([]byte(p.GetMetadata()), &parsedMeta)
	}
	displayName := parsedMeta.DisplayName
	if displayName == "" {
		displayName = p.GetName()
	}
	if displayName == "" {
		displayName = p.GetIdentity()
	}

	occ := models.VoiceOccupant{
		UserID:       p.GetIdentity(),
		DisplayName:  displayName,
		Username:     parsedMeta.Username,
		AvatarURL:    parsedMeta.AvatarURL,
		IsMuted:      false,
		JoinedAt:     now.Format(time.RFC3339Nano),
		LastActiveAt: now.Unix(),
	}

	var groupPublicID string
	if s.redis != nil {
		meta, mErr := s.redis.GetRoomMeta(ctx, channelUUID)
		if mErr == nil {
			groupPublicID = meta.GroupPublicID
		}
		meta.SessionID = sess.ID
		meta.RoomName = roomName
		if meta.StartedAt == "" {
			meta.StartedAt = now.Format(time.RFC3339Nano)
		}
		_ = s.redis.SetRoomMeta(ctx, channelUUID, meta)
		if err := s.redis.UpsertOccupant(ctx, channelUUID, groupPublicID, occ); err != nil {
			return err
		}
		if cnt, err := s.redis.CountOccupants(ctx, channelUUID); err == nil && cnt > 0 {
			_ = s.pg.UpdateSessionPeak(ctx, sess.ID, int32(cnt))
		}
	}

	_ = s.pg.InsertParticipantJoin(ctx, models.CallParticipant{
		ID:        uuid.NewString(),
		SessionID: sess.ID,
		UserID:    p.GetIdentity(),
		JoinedAt:  now,
	})

	s.updateVoiceActivity(ctx, p.GetIdentity(), "In a call")
	s.publishOccupancy(ctx, channelUUID)
	return nil
}

func (s service) onParticipantLeft(ctx context.Context, channelUUID, identity string, now time.Time) error {
	var groupPublicID string
	var sessionID string
	if s.redis != nil {
		if meta, err := s.redis.GetRoomMeta(ctx, channelUUID); err == nil {
			groupPublicID = meta.GroupPublicID
			sessionID = meta.SessionID
		}
		_ = s.redis.RemoveOccupant(ctx, channelUUID, groupPublicID, identity)
	}
	if sessionID == "" {
		if sess, err := s.pg.FindActiveSessionByChannel(ctx, channelUUID); err == nil {
			sessionID = sess.ID
		}
	}
	if sessionID != "" {
		_, _ = s.pg.CloseParticipantLeave(ctx, sessionID, identity, now, "left")
	}

	s.updateVoiceActivity(ctx, identity, "")
	s.publishOccupancy(ctx, channelUUID)
	return nil
}

func (s service) onTrackChanged(ctx context.Context, channelUUID, roomName, identity string, track *livekit.TrackInfo, published bool, now time.Time) error {
	if s.redis == nil {
		return nil
	}
	occs, err := s.redis.ListOccupants(ctx, channelUUID)
	if err != nil {
		return err
	}
	meta, _ := s.redis.GetRoomMeta(ctx, channelUUID)

	var target *models.VoiceOccupant
	for i := range occs {
		if occs[i].UserID == identity {
			target = &occs[i]
			break
		}
	}
	if target == nil {
		return nil
	}

	switch track.GetSource() {
	case livekit.TrackSource_MICROPHONE:
		target.IsMuted = !published || track.GetMuted()
	case livekit.TrackSource_CAMERA:
		if published {
			if vidCnt, err := s.redis.CountVideoPublishers(ctx, channelUUID); err == nil && vidCnt >= MaxVideoPublishers && !target.IsCameraOn && !target.IsScreenSharing {
				_ = s.lk.MuteVideoTracks(ctx, roomName, identity, true)
				return nil
			}
		}
		target.IsCameraOn = published
	case livekit.TrackSource_SCREEN_SHARE:
		if published {
			if vidCnt, err := s.redis.CountVideoPublishers(ctx, channelUUID); err == nil && vidCnt >= MaxVideoPublishers && !target.IsCameraOn && !target.IsScreenSharing {
				_ = s.lk.MuteVideoTracks(ctx, roomName, identity, true)
				return nil
			}
		}
		target.IsScreenSharing = published
	}

	target.LastActiveAt = now.Unix()
	target.IdleWarned = false
	if err := s.redis.UpsertOccupant(ctx, channelUUID, meta.GroupPublicID, *target); err != nil {
		return err
	}

	if target.IsCameraOn || target.IsScreenSharing {
		s.updateVoiceActivity(ctx, identity, "In a meeting")
	} else {
		s.updateVoiceActivity(ctx, identity, "In a call")
	}

	s.publishOccupancy(ctx, channelUUID)
	return nil
}

func (s service) onRoomFinished(ctx context.Context, channelUUID string, now time.Time) error {
	var meta models.RoomMeta
	if s.redis != nil {
		meta, _ = s.redis.GetRoomMeta(ctx, channelUUID)
		occs, _ := s.redis.ListOccupants(ctx, channelUUID)
		for _, o := range occs {
			s.updateVoiceActivity(ctx, o.UserID, "")
		}
		_ = s.redis.ClearRoom(ctx, channelUUID, meta.GroupPublicID)
	}
	sessionID := meta.SessionID
	if sessionID == "" {
		if sess, err := s.pg.FindActiveSessionByChannel(ctx, channelUUID); err == nil {
			sessionID = sess.ID
		}
	}
	if sessionID != "" {
		_ = s.pg.CloseSession(ctx, sessionID, now, 0)
	}
	if s.events != nil && meta.GroupPublicID != "" && meta.ChannelPublicID != "" {
		_ = s.events.PublishOccupancyUpdated(ctx, chatEvents.VoiceOccupancyUpdated{
			GroupID:     meta.GroupPublicID,
			ChannelID:   meta.ChannelPublicID,
			IsRecording: false,
			Occupants:   []chatEvents.VoiceOccupant{},
		})
	}
	return nil
}

func (s service) onEgressEvent(ctx context.Context, info *livekit.EgressInfo, now time.Time) error {
	status := "active"
	var endedAt *time.Time
	var fileKey string
	var fileSize int64
	var dur int32

	switch info.GetStatus() {
	case livekit.EgressStatus_EGRESS_STARTING:
		status = "starting"
	case livekit.EgressStatus_EGRESS_ACTIVE:
		status = "active"
	case livekit.EgressStatus_EGRESS_ENDING:
		status = "stopping"
	case livekit.EgressStatus_EGRESS_COMPLETE:
		status = "completed"
		endedAt = &now
		if results := info.GetFileResults(); len(results) > 0 {
			res := results[0]
			fileKey = res.GetFilename()
			fileSize = res.GetSize()
			dur = int32(res.GetDuration() / int64(time.Second))
		}
	case livekit.EgressStatus_EGRESS_FAILED, livekit.EgressStatus_EGRESS_ABORTED, livekit.EgressStatus_EGRESS_LIMIT_REACHED:
		status = "failed"
		endedAt = &now
	}

	rec, err := s.pg.UpdateRecordingByEgressID(ctx, info.GetEgressId(), status, fileKey, fileSize, dur, nil, endedAt)
	if err != nil {
		return nil
	}

	isRecording := status == "starting" || status == "active" || status == "stopping"
	if s.redis != nil {
		if meta, mErr := s.redis.GetRoomMeta(ctx, rec.ChannelID); mErr == nil {
			meta.IsRecording = isRecording
			if !isRecording {
				meta.ActiveRecordingID = ""
			}
			_ = s.redis.SetRoomMeta(ctx, rec.ChannelID, meta)
			if s.events != nil && meta.GroupPublicID != "" && meta.ChannelPublicID != "" {
				_ = s.events.PublishRecordingUpdated(ctx, chatEvents.VoiceRecordingUpdated{
					GroupID:         meta.GroupPublicID,
					ChannelID:       meta.ChannelPublicID,
					IsRecording:     isRecording,
					RecordingID:     rec.PublicID,
					StartedByUserID: rec.StartedBy,
					StartedAt:       rec.StartedAt.Format(time.RFC3339Nano),
				})
			}
		}
	}
	return nil
}

func (s service) updateVoiceActivity(ctx context.Context, userID, status string) {
	if s.userPresence == nil || userID == "" {
		return
	}
	_, _ = s.userPresence.SetVoiceActivityStatus(ctx, &userv1.SetVoiceActivityStatusRequest{
		UserId:         userID,
		ActivityStatus: status,
	})
}

func (s service) publishOccupancy(ctx context.Context, channelUUID string) {
	if s.events == nil || s.redis == nil {
		return
	}
	meta, err := s.redis.GetRoomMeta(ctx, channelUUID)
	if err != nil || meta.GroupPublicID == "" || meta.ChannelPublicID == "" {
		return
	}
	occs, err := s.redis.ListOccupants(ctx, channelUUID)
	if err != nil {
		return
	}
	evOccs := make([]chatEvents.VoiceOccupant, len(occs))
	for i, o := range occs {
		evOccs[i] = chatEvents.VoiceOccupant{
			UserID:           o.UserID,
			DisplayName:      o.DisplayName,
			Username:         o.Username,
			AvatarURL:        o.AvatarURL,
			IsMuted:          o.IsMuted,
			IsDeafened:       o.IsDeafened,
			IsServerMuted:    o.IsServerMuted,
			IsServerDeafened: o.IsServerDeafened,
			IsCameraOn:       o.IsCameraOn,
			IsScreenSharing:  o.IsScreenSharing,
			JoinedAt:         o.JoinedAt,
		}
	}
	_ = s.events.PublishOccupancyUpdated(ctx, chatEvents.VoiceOccupancyUpdated{
		GroupID:     meta.GroupPublicID,
		ChannelID:   meta.ChannelPublicID,
		IsRecording: meta.IsRecording,
		Occupants:   evOccs,
	})
}
