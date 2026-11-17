package grpc

import (
	"context"

	channelservice "github.com/wichat/wichat/backend/ms-user/internal/channel/service"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
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

func categoryPermOverrideToPB(o models.PublicChannelCategoryPermissionOverride) *userv1.ChannelCategoryPermissionOverride {
	return &userv1.ChannelCategoryPermissionOverride{
		TargetType:       o.TargetType,
		TargetId:         o.TargetID,
		TargetName:       o.TargetName,
		TargetColor:      o.TargetColor,
		TargetAvatarUrl:  o.TargetAvatarURL,
		IsEveryone:       o.IsEveryone,
		AllowPermissions: o.AllowPermissions,
		DenyPermissions:  o.DenyPermissions,
	}
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

func (s *Server) UpdateGroupChannel(ctx context.Context, req *userv1.UpdateGroupChannelRequest) (*userv1.UpdateGroupChannelResponse, error) {
	// Service
	ch, err := s.svc.UpdateGroupChannel(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId(), models.UpdateChannelInput{
		Name:                  req.Name,
		IsPrivate:             req.IsPrivate,
		SlowmodeSeconds:       req.SlowmodeSeconds,
		ThreadSlowmodeSeconds: req.ThreadSlowmodeSeconds,
		ContentVisibility:     req.ContentVisibility,
		ForumDefaultLayout:    req.ForumDefaultLayout,
		ForumSortOrder:        req.ForumSortOrder,
		ForumTagMatching:      req.ForumTagMatching,
		BitrateKbps:           req.BitrateKbps,
		VideoQuality:          req.VideoQuality,
		UserLimit:             req.UserLimit,
		RTCRegion:             req.RtcRegion,
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.UpdateGroupChannelResponse{Channel: authPb.ChannelSummary(ch)}, nil
}

func (s *Server) DeleteGroupChannel(ctx context.Context, req *userv1.DeleteGroupChannelRequest) (*userv1.DeleteGroupChannelResponse, error) {
	// Service
	if err := s.svc.DeleteGroupChannel(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.DeleteGroupChannelResponse{Ok: true}, nil
}

func (s *Server) ListChannelPermissions(ctx context.Context, req *userv1.ListChannelPermissionsRequest) (*userv1.ListChannelPermissionsResponse, error) {
	// Service
	overrides, err := s.svc.ListChannelPermissions(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}
	out := make([]*userv1.ChannelCategoryPermissionOverride, len(overrides))
	for i, o := range overrides {
		out[i] = categoryPermOverrideToPB(o)
	}

	// Response
	return &userv1.ListChannelPermissionsResponse{Overrides: out}, nil
}

func (s *Server) UpsertChannelPermission(ctx context.Context, req *userv1.UpsertChannelPermissionRequest) (*userv1.UpsertChannelPermissionResponse, error) {
	// Service
	override, err := s.svc.UpsertChannelPermission(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetChannelId(),
		req.GetTargetType(),
		req.GetTargetId(),
		req.GetAllowPermissions(),
		req.GetDenyPermissions(),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.UpsertChannelPermissionResponse{Override: categoryPermOverrideToPB(override)}, nil
}

func (s *Server) DeleteChannelPermission(ctx context.Context, req *userv1.DeleteChannelPermissionRequest) (*userv1.DeleteChannelPermissionResponse, error) {
	// Service
	if err := s.svc.DeleteChannelPermission(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetChannelId(),
		req.GetTargetType(),
		req.GetTargetId(),
	); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.DeleteChannelPermissionResponse{Ok: true}, nil
}

func (s *Server) SyncChannelPermissions(ctx context.Context, req *userv1.SyncChannelPermissionsRequest) (*userv1.SyncChannelPermissionsResponse, error) {
	// Service
	ch, err := s.svc.SyncChannelPermissions(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.SyncChannelPermissionsResponse{Channel: authPb.ChannelSummary(ch)}, nil
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
		ChannelUuid:       access.ChannelID,
		UserUuid:          access.UserID,
		Permissions:       access.Permissions,
		ChannelType:       access.ChannelType,
		SlowmodeSeconds:   access.SlowmodeSeconds,
		ContentVisibility: access.ContentVisibility,
	}, nil
}
