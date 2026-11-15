package grpc

import (
	"context"

	grouprolesvc "github.com/wichat/wichat/backend/ms-user/internal/group-role/service"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/auth-pb"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server implements userv1.GroupRoleServiceServer.
type Server struct {
	userv1.UnimplementedGroupRoleServiceServer
	svc grouprolesvc.Service
}

// NewServer creates a new GroupRoleServiceServer.
func NewServer(svc grouprolesvc.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) ListGroupRoles(ctx context.Context, req *userv1.ListGroupRolesRequest) (*userv1.ListGroupRolesResponse, error) {
	roles, hideRoles, err := s.svc.ListGroupRolesWithSettings(ctx, req.GetAccessToken(), req.GetGroupId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ListGroupRolesResponse{
		Roles:     authPb.GroupRolesPb(roles),
		HideRoles: hideRoles,
	}, nil
}

func (s *Server) CreateGroupRole(ctx context.Context, req *userv1.CreateGroupRoleRequest) (*userv1.CreateGroupRoleResponse, error) {
	role, err := s.svc.CreateGroupRole(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetName())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.CreateGroupRoleResponse{Role: authPb.GroupRolePb(*role)}, nil
}

func (s *Server) UpdateGroupRole(ctx context.Context, req *userv1.UpdateGroupRoleRequest) (*userv1.UpdateGroupRoleResponse, error) {
	var namePtr *string
	if req.GetName() != "" {
		n := req.GetName()
		namePtr = &n
	}
	var colorPtr *string
	if req.GetColor() != "" {
		c := req.GetColor()
		colorPtr = &c
	}
	h := req.GetHoist()
	m := req.GetMentionable()
	perms := req.GetPermissions()
	if req.GetClearPermissions() && perms == nil {
		perms = []string{}
	}
	patch := models.GroupRolePatch{
		Name:        namePtr,
		Color:       colorPtr,
		Hoist:       &h,
		Mentionable: &m,
		Permissions: perms,
	}

	role, err := s.svc.UpdateGroupRole(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetRoleId(), patch)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UpdateGroupRoleResponse{Role: authPb.GroupRolePb(*role)}, nil
}

func (s *Server) DuplicateGroupRole(ctx context.Context, req *userv1.DuplicateGroupRoleRequest) (*userv1.DuplicateGroupRoleResponse, error) {
	role, err := s.svc.DuplicateGroupRole(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetRoleId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.DuplicateGroupRoleResponse{Role: authPb.GroupRolePb(*role)}, nil
}

func (s *Server) DeleteGroupRole(ctx context.Context, req *userv1.DeleteGroupRoleRequest) (*userv1.DeleteGroupRoleResponse, error) {
	if err := s.svc.DeleteGroupRole(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetRoleId()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.DeleteGroupRoleResponse{Ok: true}, nil
}

func (s *Server) ListGroupRoleMembers(ctx context.Context, req *userv1.ListGroupRoleMembersRequest) (*userv1.ListGroupRoleMembersResponse, error) {
	members, err := s.svc.ListGroupRoleMembers(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetRoleId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ListGroupRoleMembersResponse{
		Members: authPb.GroupMemberEntries(members),
	}, nil
}

func (s *Server) AddGroupRoleMembers(ctx context.Context, req *userv1.AddGroupRoleMembersRequest) (*userv1.AddGroupRoleMembersResponse, error) {
	role, err := s.svc.AddGroupRoleMembers(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetRoleId(), req.GetUserIds())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.AddGroupRoleMembersResponse{Role: authPb.GroupRolePb(*role)}, nil
}

func (s *Server) RemoveGroupRoleMember(ctx context.Context, req *userv1.RemoveGroupRoleMemberRequest) (*userv1.RemoveGroupRoleMemberResponse, error) {
	role, err := s.svc.RemoveGroupRoleMember(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetRoleId(), req.GetUserId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.RemoveGroupRoleMemberResponse{Role: authPb.GroupRolePb(*role)}, nil
}

func (s *Server) UpdateGroupHideRoles(ctx context.Context, req *userv1.UpdateGroupHideRolesRequest) (*userv1.UpdateGroupHideRolesResponse, error) {
	hideRoles, err := s.svc.UpdateGroupHideRoles(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetHideRoles())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UpdateGroupHideRolesResponse{HideRoles: hideRoles}, nil
}
