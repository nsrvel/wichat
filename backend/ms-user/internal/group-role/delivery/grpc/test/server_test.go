package test

import (
	"context"
	"testing"

	grouproledelivery "github.com/wichat/wichat/backend/ms-user/internal/group-role/delivery/grpc"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestGroupRoleGRPCServer(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	roleSvc := testSupport.NewGroupRoleService(t, st)
	srv := grouproledelivery.NewServer(roleSvc)

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "owner-role-grpc@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-01",
	})
	group, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Role Group")
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}

	resp, err := srv.ListGroupRoles(ctx, &userv1.ListGroupRolesRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
	})
	if err != nil || len(resp.GetRoles()) != 2 {
		t.Fatalf("ListGroupRoles: got %d, err: %v", len(resp.GetRoles()), err)
	}

	upResp, err := srv.UpdateGroupRole(ctx, &userv1.UpdateGroupRoleRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
		RoleId:      resp.GetRoles()[1].GetId(),
		Name:        "SuperMember",
	})
	if err != nil || upResp.GetRole().GetName() != "SuperMember" {
		t.Fatalf("UpdateGroupRole: %v, resp: %+v", err, upResp)
	}

	created, err := srv.CreateGroupRole(ctx, &userv1.CreateGroupRoleRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
	})
	if err != nil || created.GetRole().GetName() != "new role" {
		t.Fatalf("CreateGroupRole: %v, resp: %+v", err, created)
	}

	dup, err := srv.DuplicateGroupRole(ctx, &userv1.DuplicateGroupRoleRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
		RoleId:      created.GetRole().GetId(),
	})
	if err != nil || dup.GetRole().GetName() != "new role Copy" {
		t.Fatalf("DuplicateGroupRole: %v, resp: %+v", err, dup)
	}

	if _, err := srv.ListGroupRoleMembers(ctx, &userv1.ListGroupRoleMembersRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
		RoleId:      created.GetRole().GetId(),
	}); err != nil {
		t.Fatalf("ListGroupRoleMembers: %v", err)
	}

	hideResp, err := srv.UpdateGroupHideRoles(ctx, &userv1.UpdateGroupHideRolesRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
		HideRoles:   true,
	})
	if err != nil || !hideResp.GetHideRoles() {
		t.Fatalf("UpdateGroupHideRoles: %v, resp: %+v", err, hideResp)
	}

	if _, err := srv.DeleteGroupRole(ctx, &userv1.DeleteGroupRoleRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
		RoleId:      dup.GetRole().GetId(),
	}); err != nil {
		t.Fatalf("DeleteGroupRole: %v", err)
	}
}
