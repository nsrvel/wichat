package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	grouproledelivery "github.com/wichat/wichat/backend/api-gateway/internal/group-role/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeGroupRole struct{}

func (fakeGroupRole) ListGroupRoles(_ context.Context, in *userv1.ListGroupRolesRequest) (*userv1.ListGroupRolesResponse, error) {
	if in.GetGroupId() == "err" {
		return nil, status.Error(codes.PermissionDenied, "error.forbidden")
	}
	return &userv1.ListGroupRolesResponse{
		Roles:     []*userv1.GroupRole{{Id: "r1", Name: "@everyone", Permissions: []string{"view_channels"}}},
		HideRoles: true,
	}, nil
}

func (fakeGroupRole) CreateGroupRole(_ context.Context, in *userv1.CreateGroupRoleRequest) (*userv1.CreateGroupRoleResponse, error) {
	name := in.GetName()
	if name == "" {
		name = "new role"
	}
	return &userv1.CreateGroupRoleResponse{Role: &userv1.GroupRole{Id: "r2", Name: name}}, nil
}

func (fakeGroupRole) UpdateGroupRole(_ context.Context, in *userv1.UpdateGroupRoleRequest) (*userv1.UpdateGroupRoleResponse, error) {
	return &userv1.UpdateGroupRoleResponse{
		Role: &userv1.GroupRole{Id: in.GetRoleId(), Name: in.GetName(), Permissions: in.GetPermissions()},
	}, nil
}

func (fakeGroupRole) DuplicateGroupRole(_ context.Context, in *userv1.DuplicateGroupRoleRequest) (*userv1.DuplicateGroupRoleResponse, error) {
	return &userv1.DuplicateGroupRoleResponse{Role: &userv1.GroupRole{Id: "r3", Name: in.GetRoleId() + " Copy"}}, nil
}

func (fakeGroupRole) DeleteGroupRole(context.Context, *userv1.DeleteGroupRoleRequest) (*userv1.DeleteGroupRoleResponse, error) {
	return &userv1.DeleteGroupRoleResponse{Ok: true}, nil
}

func (fakeGroupRole) ListGroupRoleMembers(context.Context, *userv1.ListGroupRoleMembersRequest) (*userv1.ListGroupRoleMembersResponse, error) {
	return &userv1.ListGroupRoleMembersResponse{
		Members: []*userv1.GroupMemberEntry{{UserId: "u1", DisplayName: "Alice", Username: "alice", Role: "member"}},
	}, nil
}

func (fakeGroupRole) AddGroupRoleMembers(_ context.Context, in *userv1.AddGroupRoleMembersRequest) (*userv1.AddGroupRoleMembersResponse, error) {
	return &userv1.AddGroupRoleMembersResponse{Role: &userv1.GroupRole{Id: in.GetRoleId(), MemberCount: 1}}, nil
}

func (fakeGroupRole) RemoveGroupRoleMember(_ context.Context, in *userv1.RemoveGroupRoleMemberRequest) (*userv1.RemoveGroupRoleMemberResponse, error) {
	return &userv1.RemoveGroupRoleMemberResponse{Role: &userv1.GroupRole{Id: in.GetRoleId(), MemberCount: 0}}, nil
}

func (fakeGroupRole) UpdateGroupHideRoles(_ context.Context, in *userv1.UpdateGroupHideRolesRequest) (*userv1.UpdateGroupHideRolesResponse, error) {
	return &userv1.UpdateGroupHideRolesResponse{HideRoles: in.GetHideRoles()}, nil
}

func testGroupRoleHandler() grouproledelivery.Handler {
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{GroupRole: fakeGroupRole{}}}}
	return grouproledelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))
}

func TestGroupRoleHandlerAuthAndEndpoints(t *testing.T) {
	h := testGroupRoleHandler()

	// Unauthenticated check
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/g1/roles", nil)
	rec := httptest.NewRecorder()
	h.ListGroupRoles(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	// ListGroupRoles
	req = httptest.NewRequest(http.MethodGet, "/api/v1/groups/g1/roles", nil)
	req.Header.Set("Authorization", "Bearer tok")
	req.SetPathValue("id", "g1")
	rec = httptest.NewRecorder()
	h.ListGroupRoles(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status %d: %s", rec.Code, rec.Body.String())
	}

	// CreateGroupRole
	req = httptest.NewRequest(http.MethodPost, "/api/v1/groups/g1/roles", strings.NewReader(`{"name":""}`))
	req.Header.Set("Authorization", "Bearer tok")
	req.SetPathValue("id", "g1")
	rec = httptest.NewRecorder()
	h.CreateGroupRole(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status %d: %s", rec.Code, rec.Body.String())
	}

	// UpdateGroupRole (clear permissions)
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/groups/g1/roles/r2", strings.NewReader(`{"name":"Mod","color":"#1abc9c","hoist":true,"mentionable":false,"permissions":[]}`))
	req.Header.Set("Authorization", "Bearer tok")
	req.SetPathValue("id", "g1")
	req.SetPathValue("roleId", "r2")
	rec = httptest.NewRecorder()
	h.UpdateGroupRole(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status %d: %s", rec.Code, rec.Body.String())
	}

	// DuplicateGroupRole
	req = httptest.NewRequest(http.MethodPost, "/api/v1/groups/g1/roles/r2/duplicate", nil)
	req.Header.Set("Authorization", "Bearer tok")
	req.SetPathValue("id", "g1")
	req.SetPathValue("roleId", "r2")
	rec = httptest.NewRecorder()
	h.DuplicateGroupRole(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("duplicate status %d: %s", rec.Code, rec.Body.String())
	}

	// ListGroupRoleMembers
	req = httptest.NewRequest(http.MethodGet, "/api/v1/groups/g1/roles/r2/members", nil)
	req.Header.Set("Authorization", "Bearer tok")
	req.SetPathValue("id", "g1")
	req.SetPathValue("roleId", "r2")
	rec = httptest.NewRecorder()
	h.ListGroupRoleMembers(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list members status %d: %s", rec.Code, rec.Body.String())
	}

	// AddGroupRoleMembers
	req = httptest.NewRequest(http.MethodPost, "/api/v1/groups/g1/roles/r2/members", strings.NewReader(`{"user_ids":["u1"]}`))
	req.Header.Set("Authorization", "Bearer tok")
	req.SetPathValue("id", "g1")
	req.SetPathValue("roleId", "r2")
	rec = httptest.NewRecorder()
	h.AddGroupRoleMembers(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("add members status %d: %s", rec.Code, rec.Body.String())
	}

	// RemoveGroupRoleMember
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/groups/g1/roles/r2/members/u1", nil)
	req.Header.Set("Authorization", "Bearer tok")
	req.SetPathValue("id", "g1")
	req.SetPathValue("roleId", "r2")
	req.SetPathValue("userId", "u1")
	rec = httptest.NewRecorder()
	h.RemoveGroupRoleMember(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("remove member status %d: %s", rec.Code, rec.Body.String())
	}

	// UpdateGroupHideRoles
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/groups/g1/roles/settings", strings.NewReader(`{"hide_roles":true}`))
	req.Header.Set("Authorization", "Bearer tok")
	req.SetPathValue("id", "g1")
	rec = httptest.NewRecorder()
	h.UpdateGroupHideRoles(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("hide roles status %d: %s", rec.Code, rec.Body.String())
	}
	var hideEnv struct {
		Data struct {
			HideRoles bool `json:"hide_roles"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &hideEnv); err != nil || !hideEnv.Data.HideRoles {
		t.Fatalf("unexpected hide roles payload: %s (%v)", rec.Body.String(), err)
	}

	// DeleteGroupRole
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/groups/g1/roles/r2", nil)
	req.Header.Set("Authorization", "Bearer tok")
	req.SetPathValue("id", "g1")
	req.SetPathValue("roleId", "r2")
	rec = httptest.NewRecorder()
	h.DeleteGroupRole(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status %d: %s", rec.Code, rec.Body.String())
	}
}
