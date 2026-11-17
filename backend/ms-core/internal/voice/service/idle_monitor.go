package service

import (
	"context"

	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
)

func (s service) RunIdleCheck(ctx context.Context) error {
	if s.redis == nil {
		return nil
	}
	channelUUIDs, err := s.redis.ListAllOccupiedChannelUUIDs(ctx)
	if err != nil {
		return err
	}

	now := s.now()
	nowUnix := now.Unix()
	warnCutoff := now.Add(-IdleWarningAfter).Unix()
	kickCutoff := now.Add(-IdleKickAfter).Unix()

	for _, chUUID := range channelUUIDs {
		meta, err := s.redis.GetRoomMeta(ctx, chUUID)
		if err != nil {
			continue
		}
		occs, err := s.redis.ListOccupants(ctx, chUUID)
		if err != nil {
			continue
		}
		for _, occ := range occs {
			if occ.IsCameraOn || occ.IsScreenSharing {
				continue
			}
			if occ.LastActiveAt == 0 {
				_ = s.redis.TouchOccupantActivity(ctx, chUUID, occ.UserID, nowUnix)
				continue
			}
			if occ.LastActiveAt <= kickCutoff {
				_ = s.lk.RemoveParticipant(ctx, meta.RoomName, occ.UserID)
				_ = s.redis.RemoveOccupant(ctx, chUUID, meta.GroupPublicID, occ.UserID)
				if meta.SessionID != "" {
					_, _ = s.pg.CloseParticipantLeave(ctx, meta.SessionID, occ.UserID, now, "idle_timeout")
				}
				s.updateVoiceActivity(ctx, occ.UserID, "")
				s.publishOccupancy(ctx, chUUID)
				continue
			}
			if occ.LastActiveAt <= warnCutoff && !occ.IdleWarned {
				_ = s.redis.MarkOccupantIdleWarned(ctx, chUUID, occ.UserID, true)
				if s.events != nil && meta.GroupPublicID != "" && meta.ChannelPublicID != "" {
					_ = s.events.PublishIdleWarning(ctx, chatEvents.VoiceIdleWarning{
						GroupID:          meta.GroupPublicID,
						ChannelID:        meta.ChannelPublicID,
						UserID:           occ.UserID,
						ExpiresInSeconds: 120,
					})
				}
			}
		}
	}

	return nil
}
