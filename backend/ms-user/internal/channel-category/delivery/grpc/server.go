package grpc

import (
	"context"

	categoryservice "github.com/wichat/wichat/backend/ms-user/internal/channel-category/service"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/auth-pb"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server implements userv1.ChannelCategoryServiceServer.
type Server struct {
	userv1.UnimplementedChannelCategoryServiceServer
	svc categoryservice.Service
}

// NewServer constructs channel category RPC handlers.
func NewServer(svc categoryservice.Service) *Server {
	return &Server{svc: svc}
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

func (s *Server) UpdateChannelCategory(ctx context.Context, req *userv1.UpdateChannelCategoryRequest) (*userv1.UpdateChannelCategoryResponse, error) {
	// Service
	category, err := s.svc.UpdateGroupCategory(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetCategoryId(), req.Name, req.IsPrivate)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.UpdateChannelCategoryResponse{Category: authPb.ChannelCategory(category)}, nil
}

func (s *Server) DeleteChannelCategory(ctx context.Context, req *userv1.DeleteChannelCategoryRequest) (*userv1.DeleteChannelCategoryResponse, error) {
	// Service
	if err := s.svc.DeleteGroupCategory(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetCategoryId()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.DeleteChannelCategoryResponse{Ok: true}, nil
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

func (s *Server) ListChannelCategoryPermissions(ctx context.Context, req *userv1.ListChannelCategoryPermissionsRequest) (*userv1.ListChannelCategoryPermissionsResponse, error) {
	// Service
	overrides, err := s.svc.ListCategoryPermissions(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetCategoryId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}
	out := make([]*userv1.ChannelCategoryPermissionOverride, len(overrides))
	for i, o := range overrides {
		out[i] = categoryPermOverrideToPB(o)
	}

	// Response
	return &userv1.ListChannelCategoryPermissionsResponse{Overrides: out}, nil
}

func (s *Server) UpsertChannelCategoryPermission(ctx context.Context, req *userv1.UpsertChannelCategoryPermissionRequest) (*userv1.UpsertChannelCategoryPermissionResponse, error) {
	// Service
	override, err := s.svc.UpsertCategoryPermission(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetCategoryId(),
		req.GetTargetType(),
		req.GetTargetId(),
		req.GetAllowPermissions(),
		req.GetDenyPermissions(),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.UpsertChannelCategoryPermissionResponse{Override: categoryPermOverrideToPB(override)}, nil
}

func (s *Server) DeleteChannelCategoryPermission(ctx context.Context, req *userv1.DeleteChannelCategoryPermissionRequest) (*userv1.DeleteChannelCategoryPermissionResponse, error) {
	// Service
	if err := s.svc.DeleteCategoryPermission(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetCategoryId(),
		req.GetTargetType(),
		req.GetTargetId(),
	); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.DeleteChannelCategoryPermissionResponse{Ok: true}, nil
}

func (s *Server) ReorderChannelCategories(ctx context.Context, req *userv1.ReorderChannelCategoriesRequest) (*userv1.ReorderChannelCategoriesResponse, error) {
	// Service
	if err := s.svc.ReorderChannelCategories(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetCategoryIds()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.ReorderChannelCategoriesResponse{Ok: true}, nil
}
