package grpc

import (
	"context"
	"time"

	presencesvc "github.com/wichat/wichat/backend/ms-user/internal/presence/service"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server implements userv1.PresenceServiceServer.
type Server struct {
	userv1.UnimplementedPresenceServiceServer
	svc presencesvc.Service
}

// NewServer constructs a PresenceServiceServer.
func NewServer(svc presencesvc.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) HeartbeatPresence(ctx context.Context, req *userv1.HeartbeatPresenceRequest) (*userv1.PresenceTransitionResponse, error) {
	tr, err := s.svc.HeartbeatPresence(ctx, req.GetAccessToken(), req.GetConnId(), req.GetActivity())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}
	return &userv1.PresenceTransitionResponse{
		UserId:          tr.UserID,
		PreviousStatus:  tr.PreviousStatus,
		EffectiveStatus: tr.EffectiveStatus,
		IsOnline:        tr.IsOnline,
		GroupIds:        tr.GroupIDs,
	}, nil
}

func (s *Server) DisconnectPresence(ctx context.Context, req *userv1.DisconnectPresenceRequest) (*userv1.PresenceTransitionResponse, error) {
	tr, err := s.svc.DisconnectPresence(ctx, req.GetUserId(), req.GetConnId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}
	return &userv1.PresenceTransitionResponse{
		UserId:          tr.UserID,
		PreviousStatus:  tr.PreviousStatus,
		EffectiveStatus: tr.EffectiveStatus,
		IsOnline:        tr.IsOnline,
		GroupIds:        tr.GroupIDs,
	}, nil
}

func (s *Server) SetManualPresence(ctx context.Context, req *userv1.SetManualPresenceRequest) (*userv1.SetManualPresenceResponse, error) {
	res, err := s.svc.SetManualPresence(ctx, req.GetAccessToken(), req.GetStatus(), req.GetTtlSeconds())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}
	var expiresAt string
	if res.ExpiresAt != nil {
		expiresAt = res.ExpiresAt.UTC().Format(time.RFC3339Nano)
	}
	return &userv1.SetManualPresenceResponse{
		UserId:          res.UserID,
		ManualStatus:    res.ManualStatus,
		EffectiveStatus: res.EffectiveStatus,
		IsOnline:        res.IsOnline,
		ExpiresAt:       expiresAt,
		GroupIds:        res.GroupIDs,
	}, nil
}

func (s *Server) GetMyPresence(ctx context.Context, req *userv1.GetMyPresenceRequest) (*userv1.GetMyPresenceResponse, error) {
	res, err := s.svc.GetMyPresence(ctx, req.GetAccessToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}
	var expiresAt string
	if res.ExpiresAt != nil {
		expiresAt = res.ExpiresAt.UTC().Format(time.RFC3339Nano)
	}
	return &userv1.GetMyPresenceResponse{
		ManualStatus:    res.ManualStatus,
		EffectiveStatus: res.EffectiveStatus,
		IsOnline:        res.IsOnline,
		ExpiresAt:       expiresAt,
	}, nil
}

func (s *Server) SetVoiceActivityStatus(ctx context.Context, req *userv1.SetVoiceActivityStatusRequest) (*userv1.SetVoiceActivityStatusResponse, error) {
	groupIDs, err := s.svc.SetVoiceActivityStatus(ctx, req.GetUserId(), req.GetActivityStatus())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}
	return &userv1.SetVoiceActivityStatusResponse{
		Ok:       true,
		GroupIds: groupIDs,
	}, nil
}
