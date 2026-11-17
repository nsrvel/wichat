package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/wichat/wichat/backend/ms-core/internal/models"
	voicerepo "github.com/wichat/wichat/backend/ms-core/internal/voice/repository"
	livekitclient "github.com/wichat/wichat/backend/ms-core/pkg/livekit-client"
	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func (s service) StartRecording(ctx context.Context, accessToken, groupPublicID, channelPublicID string) (models.CallRecording, error) {
	acc, err := s.userChannel.AssertGroupChannelAccess(ctx, &userv1.AssertGroupChannelAccessRequest{
		AccessToken: accessToken,
		GroupId:     groupPublicID,
		ChannelId:   channelPublicID,
	})
	if err != nil {
		return models.CallRecording{}, err
	}

	perms := acc.GetPermissions()
	if len(perms) > 0 && !hasAnyPerm(perms, "administrator", "record_call") {
		return models.CallRecording{}, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	channelUUID := acc.GetChannelUuid()
	if _, err := s.pg.FindActiveRecordingByChannel(ctx, channelUUID); err == nil {
		return models.CallRecording{}, exception.Conflict(exception.MsgVoiceRecordingAlready, nil)
	} else if !errors.Is(err, voicerepo.ErrNotFound) {
		return models.CallRecording{}, err
	}

	now := s.now()
	roomName := roomNameForChannel(acc.GetChannelType(), channelUUID)
	sess, err := s.pg.UpsertActiveSession(ctx, models.CallSession{
		ID:               uuid.NewString(),
		PublicID:         newPublicID24(),
		ChannelID:        channelUUID,
		RoomName:         roomName,
		StartedAt:        now,
		PeakParticipants: 1,
	})
	if err != nil {
		return models.CallRecording{}, err
	}

	fileKey := fmt.Sprintf("recordings/%s/%s/%s/%d.mp4", groupPublicID, channelPublicID, sess.PublicID, now.Unix())
	s3Endpoint := strings.TrimSpace(s.opts.MinioEndpoint)
	if s3Endpoint != "" && !strings.HasPrefix(s3Endpoint, "http://") && !strings.HasPrefix(s3Endpoint, "https://") {
		if s.opts.MinioUseSSL {
			s3Endpoint = "https://" + s3Endpoint
		} else {
			s3Endpoint = "http://" + s3Endpoint
		}
	}

	egressID, err := s.lk.StartRoomCompositeEgress(ctx, roomName, livekitclient.EgressS3Config{
		Endpoint: s3Endpoint,
		Bucket:   s.opts.MinioBucket,
		Region:   s.opts.MinioRegion,
		Key:      s.opts.MinioAccessKey,
		Secret:   s.opts.MinioSecretKey,
		Filepath: fileKey,
	})
	if err != nil {
		return models.CallRecording{}, err
	}

	rec, err := s.pg.CreateRecording(ctx, models.CallRecording{
		ID:        uuid.NewString(),
		PublicID:  newPublicID24(),
		SessionID: sess.ID,
		ChannelID: channelUUID,
		EgressID:  egressID,
		Status:    "starting",
		FileKey:   fileKey,
		StartedBy: acc.GetUserUuid(),
		StartedAt: now,
	})
	if err != nil {
		return models.CallRecording{}, err
	}

	if s.redis != nil {
		meta, _ := s.redis.GetRoomMeta(ctx, channelUUID)
		meta.SessionID = sess.ID
		meta.GroupPublicID = groupPublicID
		meta.ChannelPublicID = channelPublicID
		meta.RoomName = roomName
		meta.IsRecording = true
		meta.ActiveRecordingID = rec.PublicID
		_ = s.redis.SetRoomMeta(ctx, channelUUID, meta)
	}

	if s.events != nil {
		_ = s.events.PublishRecordingUpdated(ctx, chatEvents.VoiceRecordingUpdated{
			GroupID:         groupPublicID,
			ChannelID:       channelPublicID,
			IsRecording:     true,
			RecordingID:     rec.PublicID,
			StartedByUserID: rec.StartedBy,
			StartedAt:       rec.StartedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
		})
	}
	return rec, nil
}

func (s service) StopRecording(ctx context.Context, accessToken, groupPublicID, channelPublicID, _ string) (models.CallRecording, error) {
	acc, err := s.userChannel.AssertGroupChannelAccess(ctx, &userv1.AssertGroupChannelAccessRequest{
		AccessToken: accessToken,
		GroupId:     groupPublicID,
		ChannelId:   channelPublicID,
	})
	if err != nil {
		return models.CallRecording{}, err
	}

	perms := acc.GetPermissions()
	if len(perms) > 0 && !hasAnyPerm(perms, "administrator", "record_call") {
		return models.CallRecording{}, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	active, err := s.pg.FindActiveRecordingByChannel(ctx, acc.GetChannelUuid())
	if err != nil {
		if errors.Is(err, voicerepo.ErrNotFound) {
			return models.CallRecording{}, exception.NotFound(exception.MsgVoiceRecordingNotActive, nil)
		}
		return models.CallRecording{}, err
	}

	if err := s.lk.StopEgress(ctx, active.EgressID); err != nil {
		return models.CallRecording{}, err
	}

	uid := acc.GetUserUuid()
	updated, err := s.pg.UpdateRecordingByEgressID(ctx, active.EgressID, "stopping", "", 0, 0, &uid, nil)
	if err != nil {
		return models.CallRecording{}, err
	}

	return updated, nil
}

func (s service) ListChannelRecordings(ctx context.Context, accessToken, groupPublicID, channelPublicID string, limit int) ([]models.CallRecording, error) {
	acc, err := s.userChannel.AssertGroupChannelAccess(ctx, &userv1.AssertGroupChannelAccessRequest{
		AccessToken: accessToken,
		GroupId:     groupPublicID,
		ChannelId:   channelPublicID,
	})
	if err != nil {
		return nil, err
	}

	perms := acc.GetPermissions()
	if len(perms) > 0 && !hasAnyPerm(perms, "administrator", "read_message_history", "record_call") {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	recs, err := s.pg.ListRecordingsByChannel(ctx, acc.GetChannelUuid(), limit)
	if err != nil {
		return nil, err
	}
	if s.minio != nil {
		for i := range recs {
			if recs[i].FileKey != "" && recs[i].Status == "completed" {
				recs[i].DownloadURL = s.minio.RecordingURL(recs[i].FileKey)
			}
		}
	}
	return recs, nil
}
