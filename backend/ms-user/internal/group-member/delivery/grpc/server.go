package grpc

import (
	"context"

	groupmembersvc "github.com/wichat/wichat/backend/ms-user/internal/group-member/service"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/auth-pb"
	inviteCursor "github.com/wichat/wichat/backend/ms-user/pkg/invite"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server implements userv1.GroupMemberServiceServer.
type Server struct {
	userv1.UnimplementedGroupMemberServiceServer
	svc groupmembersvc.Service
}

// NewServer creates a new GroupMemberServiceServer.
func NewServer(svc groupmembersvc.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) JoinGroup(ctx context.Context, req *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error) {
	if err := s.svc.JoinGroup(ctx, req.GetAccessToken(), req.GetGroupId()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.JoinGroupResponse{Ok: true}, nil
}

func (s *Server) LeaveGroup(ctx context.Context, req *userv1.LeaveGroupRequest) (*userv1.LeaveGroupResponse, error) {
	if err := s.svc.LeaveGroup(ctx, req.GetAccessToken(), req.GetGroupId()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.LeaveGroupResponse{Ok: true}, nil
}

func (s *Server) ListGroupMembers(ctx context.Context, req *userv1.ListGroupMembersRequest) (*userv1.ListGroupMembersResponse, error) {
	cursor, err := inviteCursor.DecodeMemberCursor(req.GetCursor())
	if err != nil {
		return nil, responsegrpc.ToStatus(exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "cursor"}))
	}

	page, err := s.svc.ListGroupMembers(ctx, req.GetAccessToken(), req.GetGroupId(), int(req.GetLimit()), cursor)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ListGroupMembersResponse{
		Members:    authPb.GroupMemberEntries(page.Members),
		NextCursor: inviteCursor.EncodeMemberCursor(page.NextCursor),
		Total:      int32(page.Total), //nolint:gosec // G115 — member totals are bounded by list limits
	}, nil
}

func (s *Server) KickGroupMember(ctx context.Context, req *userv1.KickGroupMemberRequest) (*userv1.KickGroupMemberResponse, error) {
	err := s.svc.KickGroupMember(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetUserId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.KickGroupMemberResponse{Success: true}, nil
}

func (s *Server) BanGroupMembers(ctx context.Context, req *userv1.BanGroupMembersRequest) (*userv1.BanGroupMembersResponse, error) {
	bannedCount, err := s.svc.BanGroupMembers(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetUserIds(),
		req.GetReason(),
		int(req.GetDeleteMessageSeconds()),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.BanGroupMembersResponse{BannedCount: int32(bannedCount)}, nil //nolint:gosec // G115
}
