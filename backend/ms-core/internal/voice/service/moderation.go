package service

import (
	"context"

	"github.com/wichat/wichat/backend/ms-core/internal/models"
	livekitclient "github.com/wichat/wichat/backend/ms-core/pkg/livekit-client"
	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func (s service) ModerateParticipant(ctx context.Context, in *models.ModerateParticipantInput) (models.VoiceOccupant, error) {
	acc, err := s.userChannel.AssertGroupChannelAccess(ctx, &userv1.AssertGroupChannelAccessRequest{
		AccessToken: in.AccessToken,
		GroupId:     in.GroupPublicID,
		ChannelId:   in.ChannelPublicID,
	})
	if err != nil {
		return models.VoiceOccupant{}, err
	}

	perms := acc.GetPermissions()
	switch in.Action {
	case "mute", "unmute":
		if len(perms) > 0 && !hasAnyPerm(perms, "administrator", "mute_members") {
			return models.VoiceOccupant{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
	case "deafen", "undeafen":
		if len(perms) > 0 && !hasAnyPerm(perms, "administrator", "deafen_members") {
			return models.VoiceOccupant{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
	case "disconnect":
		if len(perms) > 0 && !hasAnyPerm(perms, "administrator", "move_members", "kick_members") {
			return models.VoiceOccupant{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
	default:
		return models.VoiceOccupant{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "action"})
	}

	channelUUID := acc.GetChannelUuid()
	roomName := roomNameForChannel(acc.GetChannelType(), channelUUID)
	if in.Action == "disconnect" {
		if err := s.lk.RemoveParticipant(ctx, roomName, in.TargetUserID); err != nil {
			return models.VoiceOccupant{}, err
		}
		if s.redis != nil {
			_ = s.redis.RemoveOccupant(ctx, channelUUID, in.GroupPublicID, in.TargetUserID)
		}
		if sess, sErr := s.pg.FindActiveSessionByChannel(ctx, channelUUID); sErr == nil {
			_, _ = s.pg.CloseParticipantLeave(ctx, sess.ID, in.TargetUserID, s.now(), "kicked")
		}
		s.updateVoiceActivity(ctx, in.TargetUserID, "")
		s.publishOccupancy(ctx, channelUUID)
		if s.events != nil {
			_ = s.events.PublishModerationApplied(ctx, chatEvents.VoiceModerationApplied{
				GroupID:      in.GroupPublicID,
				ChannelID:    in.ChannelPublicID,
				TargetUserID: in.TargetUserID,
				Action:       in.Action,
				ActorUserID:  acc.GetUserUuid(),
			})
		}
		return models.VoiceOccupant{UserID: in.TargetUserID}, nil
	}

	var occ models.VoiceOccupant
	if s.redis != nil {
		occs, _ := s.redis.ListOccupants(ctx, channelUUID)
		for _, o := range occs {
			if o.UserID == in.TargetUserID {
				occ = o
				break
			}
		}
	}
	occ.UserID = in.TargetUserID

	switch in.Action {
	case "mute":
		occ.IsServerMuted = true
		_ = s.lk.MuteAudioTracks(ctx, roomName, in.TargetUserID, true)
	case "unmute":
		occ.IsServerMuted = false
	case "deafen":
		occ.IsServerDeafened = true
		occ.IsServerMuted = true
		_ = s.lk.MuteAudioTracks(ctx, roomName, in.TargetUserID, true)
	case "undeafen":
		occ.IsServerDeafened = false
	}

	if err := s.lk.UpdateParticipantPermissions(ctx, livekitclient.ParticipantModeration{
		RoomName:     roomName,
		Identity:     in.TargetUserID,
		CanPublish:   !occ.IsServerMuted && !occ.IsServerDeafened,
		CanSubscribe: !occ.IsServerDeafened,
	}); err != nil {
		return models.VoiceOccupant{}, err
	}

	if s.redis != nil {
		_ = s.redis.UpsertOccupant(ctx, channelUUID, in.GroupPublicID, occ)
	}
	if sess, sErr := s.pg.FindActiveSessionByChannel(ctx, channelUUID); sErr == nil {
		_ = s.pg.UpdateParticipantModeration(ctx, sess.ID, in.TargetUserID, &occ.IsServerMuted, &occ.IsServerDeafened)
	}

	s.publishOccupancy(ctx, channelUUID)
	if s.events != nil {
		_ = s.events.PublishModerationApplied(ctx, chatEvents.VoiceModerationApplied{
			GroupID:      in.GroupPublicID,
			ChannelID:    in.ChannelPublicID,
			TargetUserID: in.TargetUserID,
			Action:       in.Action,
			ActorUserID:  acc.GetUserUuid(),
		})
	}
	return occ, nil
}
