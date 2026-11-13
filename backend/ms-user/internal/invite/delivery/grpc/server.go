package grpc

import (
	"context"
	"time"

	invitesvc "github.com/wichat/wichat/backend/ms-user/internal/invite/service"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server implements userv1.InviteServiceServer.
type Server struct {
	userv1.UnimplementedInviteServiceServer
	svc invitesvc.Service
}

// NewServer creates a new InviteServiceServer.
func NewServer(svc invitesvc.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) GetGroupInviteLink(ctx context.Context, req *userv1.GetGroupInviteLinkRequest) (*userv1.GetGroupInviteLinkResponse, error) {
	link, err := s.svc.GetInviteLink(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.GetGroupInviteLinkResponse{
		Token:     link.Token,
		ExpiresAt: link.ExpiresAt.UTC().Format(time.RFC3339Nano),
	}, nil
}

func (s *Server) SendGroupInvites(ctx context.Context, req *userv1.SendGroupInvitesRequest) (*userv1.SendGroupInvitesResponse, error) {
	out, err := s.svc.SendGroupInvites(ctx, req.GetAccessToken(), models.SendGroupInvitesInput{
		GroupPublicID:   req.GetGroupId(),
		ChannelPublicID: req.GetChannelId(),
		UserIDs:         req.GetUserIds(),
		Emails:          req.GetEmails(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.SendGroupInvitesResponse{
		Added:   int32(out.Added),   //nolint:gosec // G115 — batch capped by MaxInviteBatchSize
		Emailed: int32(out.Emailed), //nolint:gosec
		Skipped: int32(out.Skipped), //nolint:gosec
	}, nil
}

func (s *Server) PreviewGroupInvite(ctx context.Context, req *userv1.PreviewGroupInviteRequest) (*userv1.PreviewGroupInviteResponse, error) {
	out, err := s.svc.PreviewInvite(ctx, req.GetToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	resp := &userv1.PreviewGroupInviteResponse{
		Valid:        out.Valid,
		GroupName:    out.GroupName,
		ChannelName:  out.ChannelName,
		GroupLogoUrl: out.GroupLogoURL,
		MemberCount:  int32(out.MemberCount), //nolint:gosec
		OnlineCount:  int32(out.OnlineCount), //nolint:gosec
		Reason:       out.Reason,
	}
	if !out.ExpiresAt.IsZero() {
		resp.ExpiresAt = out.ExpiresAt.UTC().Format(time.RFC3339Nano)
	}

	return resp, nil
}

func (s *Server) RedeemGroupInvite(ctx context.Context, req *userv1.RedeemGroupInviteRequest) (*userv1.RedeemGroupInviteResponse, error) {
	out, err := s.svc.RedeemInvite(ctx, req.GetAccessToken(), req.GetToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.RedeemGroupInviteResponse{
		GroupId:   out.GroupPublicID,
		ChannelId: out.ChannelPublicID,
	}, nil
}
