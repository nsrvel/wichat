package test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
)

func TestGroupRoleService(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	memberSvc := testSupport.NewGroupMemberService(t, st)
	roleSvc := testSupport.NewGroupRoleService(t, st)

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "owner-role@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-01",
	})
	member := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "member-role@example.com", Password: "correct-horse-battery", DisplayName: "Member", BirthDate: "2000-01-01",
	})

	group, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Role Group")
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if err := memberSvc.JoinGroup(ctx, member.AccessToken, group.ID); err != nil {
		t.Fatalf("JoinGroup: %v", err)
	}

	roles, err := roleSvc.ListGroupRoles(ctx, owner.AccessToken, group.ID)
	if err != nil || len(roles) != 2 {
		t.Fatalf("ListGroupRoles: got %d, err: %v", len(roles), err)
	}

	name := "Admin"
	if _, err := roleSvc.UpdateGroupRole(ctx, member.AccessToken, group.ID, roles[1].ID, models.GroupRolePatch{Name: &name}); err == nil {
		t.Fatalf("expected permission denied for member updating role")
	}

	updated, err := roleSvc.UpdateGroupRole(ctx, owner.AccessToken, group.ID, roles[1].ID, models.GroupRolePatch{
		Name:        &name,
		Permissions: []string{constants.PermManageRoles},
	})
	if err != nil || updated.Name != "Admin" {
		t.Fatalf("UpdateGroupRole owner: %v, updated: %+v", err, updated)
	}

	added, err := roleSvc.AddGroupRoleMembers(ctx, owner.AccessToken, group.ID, roles[1].ID, []string{member.User.ID})
	if err != nil || added == nil {
		t.Fatalf("AddGroupRoleMembers: %v", err)
	}

	removed, err := roleSvc.RemoveGroupRoleMember(ctx, owner.AccessToken, group.ID, roles[1].ID, member.User.ID)
	if err != nil || removed == nil {
		t.Fatalf("RemoveGroupRoleMember: %v", err)
	}

	perms, err := roleSvc.GetEffectivePermissions(ctx, st.Groups[0].ID, owner.User.ID)
	if err != nil || len(perms) != len(constants.AllKnownPermissions) {
		t.Fatalf("GetEffectivePermissions owner: %v, count %d", err, len(perms))
	}
}
