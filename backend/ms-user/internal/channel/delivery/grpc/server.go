package grpc

import (
	"context"

	channelservice "github.com/wichat/wichat/backend/ms-user/internal/channel/service"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/auth-pb"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server implements userv1.ChannelServiceServer.
type Server struct {
	userv1.UnimplementedChannelServiceServer
	svc channelservice.Service
}

// NewServer constructs channel RPC handlers.
func NewServer(svc channelservice.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) ListGroupChannels(ctx context.Context, req *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error) {
	// Service
	channels, categories, err := s.svc.ListGroupChannels(ctx, req.GetAccessToken(), req.GetGroupId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.ListGroupChannelsResponse{
		Channels:   authPb.ChannelSummaries(channels),
		Categories: authPb.ChannelCategories(categories),
	}, nil
}

func (s *Server) CreateChannelCategory(ctx context.Context, req *userv1.CreateChannelCategoryRequest) (*userv1.CreateChannelCategoryResponse, error) {
	// Service
	category, err := s.svc.CreateGroupCategory(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetName(), req.GetIsPrivate())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.CreateChannelCategoryResponse{Category: authPb.ChannelCategory(category)}, nil
}

func (s *Server) CreateGroupChannel(ctx context.Context, req *userv1.CreateGroupChannelRequest) (*userv1.CreateGroupChannelResponse, error) {
	// Service
	ch, err := s.svc.CreateGroupChannel(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetName(),
		req.GetType(),
		req.GetCategoryId(),
		req.GetIsPrivate(),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.CreateGroupChannelResponse{Channel: authPb.ChannelSummary(ch)}, nil
}

func (s *Server) ReorderChannelCategories(ctx context.Context, req *userv1.ReorderChannelCategoriesRequest) (*userv1.ReorderChannelCategoriesResponse, error) {
	// Service
	if err := s.svc.ReorderChannelCategories(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetCategoryIds()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.ReorderChannelCategoriesResponse{Ok: true}, nil
}

func (s *Server) ReorderGroupChannels(ctx context.Context, req *userv1.ReorderGroupChannelsRequest) (*userv1.ReorderGroupChannelsResponse, error) {
	// Service
	if err := s.svc.ReorderGroupChannels(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetSidebarSection(), req.GetCategoryId(), req.GetChannelIds()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.ReorderGroupChannelsResponse{Ok: true}, nil
}

func (s *Server) AssertGroupChannelAccess(ctx context.Context, req *userv1.AssertGroupChannelAccessRequest) (*userv1.AssertGroupChannelAccessResponse, error) {
	// Service
	access, err := s.svc.AssertGroupChannelAccess(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.AssertGroupChannelAccessResponse{
		ChannelUuid: access.ChannelID,
		UserUuid:    access.UserID,
		Permissions: access.Permissions,
	}, nil
}
