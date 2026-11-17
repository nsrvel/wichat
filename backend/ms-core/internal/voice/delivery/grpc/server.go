package grpc

import (
	"context"
	"time"

	"github.com/wichat/wichat/backend/ms-core/internal/models"
	voiceservice "github.com/wichat/wichat/backend/ms-core/internal/voice/service"
	corev1 "github.com/wichat/wichat/backend/wi-shared/models/gen/core/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server implements corev1.VoiceServiceServer.
type Server struct {
	corev1.UnimplementedVoiceServiceServer
	svc voiceservice.Service
}

// NewServer constructs voice RPC handlers.
func NewServer(svc voiceservice.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) IssueVoiceToken(ctx context.Context, req *corev1.IssueVoiceTokenRequest) (*corev1.IssueVoiceTokenResponse, error) {
	// Service
	res, err := s.svc.IssueVoiceToken(ctx, &models.IssueVoiceTokenInput{
		AccessToken:     req.GetAccessToken(),
		GroupPublicID:   req.GetGroupId(),
		ChannelPublicID: req.GetChannelId(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &corev1.IssueVoiceTokenResponse{
		Token:              res.Token,
		LivekitUrl:         res.LiveKitURL,
		RoomName:           res.RoomName,
		ChannelId:          res.ChannelID,
		ExpiresAt:          res.ExpiresAt.UTC().Format(time.RFC3339),
		InitialServerMuted: res.InitialServerMuted,
		CanPublishAudio:    res.CanPublishAudio,
		CanPublishVideo:    res.CanPublishVideo,
		UseVoiceActivity:   res.UseVoiceActivity,
		BitrateKbps:        res.BitrateKbps,
		VideoQuality:       res.VideoQuality,
		IsRecording:        res.IsRecording,
		Occupants:          occupantsToPB(res.Occupants),
	}, nil
}

func (s *Server) ModerateParticipant(ctx context.Context, req *corev1.ModerateParticipantRequest) (*corev1.ModerateParticipantResponse, error) {
	// Service
	occ, err := s.svc.ModerateParticipant(ctx, &models.ModerateParticipantInput{
		AccessToken:     req.GetAccessToken(),
		GroupPublicID:   req.GetGroupId(),
		ChannelPublicID: req.GetChannelId(),
		TargetUserID:    req.GetTargetUserId(),
		Action:          req.GetAction(),
		Reason:          req.GetReason(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &corev1.ModerateParticipantResponse{
		Ok:       true,
		Occupant: occupantToPB(occ),
	}, nil
}

func (s *Server) HandleWebhook(ctx context.Context, req *corev1.HandleWebhookRequest) (*corev1.HandleWebhookResponse, error) {
	// Service
	if err := s.svc.HandleWebhook(ctx, req.GetAuthHeader(), req.GetRawBody()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &corev1.HandleWebhookResponse{Ok: true}, nil
}

func (s *Server) ListGroupOccupancy(ctx context.Context, req *corev1.ListGroupOccupancyRequest) (*corev1.ListGroupOccupancyResponse, error) {
	// Service
	channels, err := s.svc.ListGroupOccupancy(ctx, req.GetAccessToken(), req.GetGroupId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	out := make([]*corev1.ChannelOccupancy, 0, len(channels))
	for _, ch := range channels {
		out = append(out, &corev1.ChannelOccupancy{
			ChannelId:   ch.ChannelID,
			IsRecording: ch.IsRecording,
			Occupants:   occupantsToPB(ch.Occupants),
		})
	}
	return &corev1.ListGroupOccupancyResponse{Channels: out}, nil
}

func (s *Server) StartRecording(ctx context.Context, req *corev1.StartRecordingRequest) (*corev1.StartRecordingResponse, error) {
	// Service
	rec, err := s.svc.StartRecording(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &corev1.StartRecordingResponse{Recording: recordingToPB(rec)}, nil
}

func (s *Server) StopRecording(ctx context.Context, req *corev1.StopRecordingRequest) (*corev1.StopRecordingResponse, error) {
	// Service
	rec, err := s.svc.StopRecording(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId(), req.GetRecordingId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &corev1.StopRecordingResponse{Recording: recordingToPB(rec)}, nil
}

func (s *Server) ListChannelRecordings(ctx context.Context, req *corev1.ListChannelRecordingsRequest) (*corev1.ListChannelRecordingsResponse, error) {
	// Service
	recs, err := s.svc.ListChannelRecordings(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId(), int(req.GetLimit()))
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	out := make([]*corev1.CallRecordingSummary, 0, len(recs))
	for _, rec := range recs {
		out = append(out, recordingToPB(rec))
	}
	return &corev1.ListChannelRecordingsResponse{Recordings: out}, nil
}

func (s *Server) AcknowledgeIdleWarning(ctx context.Context, req *corev1.AcknowledgeIdleWarningRequest) (*corev1.AcknowledgeIdleWarningResponse, error) {
	// Service
	if err := s.svc.AcknowledgeIdleWarning(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &corev1.AcknowledgeIdleWarningResponse{Ok: true}, nil
}

func (s *Server) GetQuotaUsage(ctx context.Context, req *corev1.GetQuotaUsageRequest) (*corev1.GetQuotaUsageResponse, error) {
	// Service
	res, err := s.svc.GetQuotaUsage(ctx, req.GetAccessToken(), req.GetGroupId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &corev1.GetQuotaUsageResponse{
		Month:                   res.Month,
		UsedParticipantMinutes:  res.UsedParticipantMinutes,
		LimitParticipantMinutes: res.LimitParticipantMinutes,
		UsagePercent:            res.UsagePercent,
		WarningActive:           res.WarningActive,
	}, nil
}

func occupantsToPB(in []models.VoiceOccupant) []*corev1.VoiceOccupant {
	out := make([]*corev1.VoiceOccupant, 0, len(in))
	for _, occ := range in {
		out = append(out, occupantToPB(occ))
	}
	return out
}

func occupantToPB(occ models.VoiceOccupant) *corev1.VoiceOccupant {
	return &corev1.VoiceOccupant{
		UserId:           occ.UserID,
		DisplayName:      occ.DisplayName,
		Username:         occ.Username,
		AvatarUrl:        occ.AvatarURL,
		IsMuted:          occ.IsMuted,
		IsDeafened:       occ.IsDeafened,
		IsServerMuted:    occ.IsServerMuted,
		IsServerDeafened: occ.IsServerDeafened,
		IsCameraOn:       occ.IsCameraOn,
		IsScreenSharing:  occ.IsScreenSharing,
		JoinedAt:         occ.JoinedAt,
	}
}

func recordingToPB(rec models.CallRecording) *corev1.CallRecordingSummary {
	stoppedBy := ""
	if rec.StoppedBy != nil {
		stoppedBy = *rec.StoppedBy
	}
	endedAt := ""
	if rec.EndedAt != nil {
		endedAt = rec.EndedAt.UTC().Format(time.RFC3339)
	}
	id := rec.PublicID
	if id == "" {
		id = rec.ID
	}
	return &corev1.CallRecordingSummary{
		Id:              id,
		SessionId:       rec.SessionID,
		ChannelId:       rec.ChannelID,
		EgressId:        rec.EgressID,
		Status:          rec.Status,
		FileKey:         rec.FileKey,
		FileSizeBytes:   rec.FileSizeBytes,
		DurationSeconds: rec.DurationSeconds,
		StartedBy:       rec.StartedBy,
		StoppedBy:       stoppedBy,
		StartedAt:       rec.StartedAt.UTC().Format(time.RFC3339),
		EndedAt:         endedAt,
		DownloadUrl:     rec.DownloadURL,
	}
}
