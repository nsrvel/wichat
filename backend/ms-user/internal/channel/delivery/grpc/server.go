package grpc

import (
	"context"

	channelservice "github.com/wichat/wichat/backend/ms-user/internal/channel/service"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/converter"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server is channel metadata gRPC delivery (still registered on AuthService until proto splits).
type Server struct {
	svc channelservice.Service
}

// NewServer constructs channel RPC handlers.
func NewServer(svc channelservice.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) ListGroupChannels(ctx context.Context, req *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error) {
	// Service
	channels, err := s.svc.ListGroupChannels(ctx, req.GetAccessToken(), req.GetGroupId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.ListGroupChannelsResponse{Channels: authPb.ChannelSummaries(channels)}, nil
}
