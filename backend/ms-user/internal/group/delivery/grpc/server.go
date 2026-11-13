package grpc

import (
	"bytes"
	"context"

	groupsvc "github.com/wichat/wichat/backend/ms-user/internal/group/service"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/auth-pb"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server implements userv1.GroupServiceServer.
type Server struct {
	userv1.UnimplementedGroupServiceServer
	svc groupsvc.Service
}

// NewServer creates a new GroupServiceServer.
func NewServer(svc groupsvc.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) ListMyGroups(ctx context.Context, req *userv1.ListMyGroupsRequest) (*userv1.ListMyGroupsResponse, error) {
	groups, err := s.svc.ListMyGroups(ctx, req.GetAccessToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ListMyGroupsResponse{Groups: authPb.GroupSummaries(groups)}, nil
}

func (s *Server) ListJoinableGroups(ctx context.Context, req *userv1.ListJoinableGroupsRequest) (*userv1.ListJoinableGroupsResponse, error) {
	groups, err := s.svc.ListJoinableGroups(ctx, req.GetAccessToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ListJoinableGroupsResponse{Groups: authPb.GroupSummaries(groups)}, nil
}

func (s *Server) CreateGroup(ctx context.Context, req *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error) {
	group, err := s.svc.CreateGroup(ctx, req.GetAccessToken(), req.GetName())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.CreateGroupResponse{Group: authPb.GroupSummary(*group)}, nil
}

func (s *Server) ReorderMyGroups(ctx context.Context, req *userv1.ReorderMyGroupsRequest) (*userv1.ReorderMyGroupsResponse, error) {
	if err := s.svc.ReorderMyGroups(ctx, req.GetAccessToken(), req.GetGroupIds()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ReorderMyGroupsResponse{Ok: true}, nil
}

func (s *Server) DeleteGroup(ctx context.Context, req *userv1.DeleteGroupRequest) (*userv1.DeleteGroupResponse, error) {
	if err := s.svc.DeleteGroup(ctx, req.GetAccessToken(), req.GetGroupId()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.DeleteGroupResponse{Ok: true}, nil
}

func (s *Server) UploadGroupIcon(ctx context.Context, req *userv1.UploadGroupIconRequest) (*userv1.UploadGroupIconResponse, error) {
	group, err := s.svc.UploadGroupIcon(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetContentType(),
		bytes.NewReader(req.GetData()),
		int64(len(req.GetData())),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UploadGroupIconResponse{Group: authPb.GroupSummary(*group)}, nil
}

func (s *Server) UpdateGroupProfile(ctx context.Context, req *userv1.UpdateGroupProfileRequest) (*userv1.UpdateGroupProfileResponse, error) {
	group, err := s.svc.UpdateGroupProfile(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetName(),
		req.GetDescription(),
		req.GetIsPrivate(),
		req.GetBannerColor(),
		req.GetClearLogo(),
		req.GetClearBannerImage(),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UpdateGroupProfileResponse{Group: authPb.GroupSummary(*group)}, nil
}

func (s *Server) UploadGroupBanner(ctx context.Context, req *userv1.UploadGroupBannerRequest) (*userv1.UploadGroupBannerResponse, error) {
	group, err := s.svc.UploadGroupBanner(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetContentType(),
		bytes.NewReader(req.GetData()),
		int64(len(req.GetData())),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UploadGroupBannerResponse{Group: authPb.GroupSummary(*group)}, nil
}
