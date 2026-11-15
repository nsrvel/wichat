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

	created, err := roleSvc.CreateGroupRole(ctx, owner.AccessToken, group.ID, "")
	if err != nil || created.Name != "new role" || len(created.Permissions) != 0 {
		t.Fatalf("CreateGroupRole default: %+v, err: %v", created, err)
	}

	if _, err := roleSvc.CreateGroupRole(ctx, member.AccessToken, group.ID, "Nope"); err == nil {
		t.Fatal("expected permission denied for member creating role")
	}

	dup, err := roleSvc.DuplicateGroupRole(ctx, owner.AccessToken, group.ID, created.PublicID)
	if err != nil || dup.Name != "new role Copy" {
		t.Fatalf("DuplicateGroupRole: %+v, err: %v", dup, err)
	}

	if _, err := roleSvc.AddGroupRoleMembers(ctx, owner.AccessToken, group.ID, created.PublicID, []string{member.User.ID}); err != nil {
		t.Fatalf("AddGroupRoleMembers custom: %v", err)
	}

	roleMembers, err := roleSvc.ListGroupRoleMembers(ctx, owner.AccessToken, group.ID, created.PublicID)
	if err != nil || len(roleMembers) != 1 || roleMembers[0].UserID != member.User.ID {
		t.Fatalf("ListGroupRoleMembers: %+v, err: %v", roleMembers, err)
	}

	hide, err := roleSvc.UpdateGroupHideRoles(ctx, owner.AccessToken, group.ID, true)
	if err != nil || !hide {
		t.Fatalf("UpdateGroupHideRoles: %v, %v", hide, err)
	}

	_, listedHide, err := roleSvc.ListGroupRolesWithSettings(ctx, member.AccessToken, group.ID)
	if err != nil || !listedHide {
		t.Fatalf("ListGroupRolesWithSettings: hide=%v, err=%v", listedHide, err)
	}

	if err := roleSvc.DeleteGroupRole(ctx, owner.AccessToken, group.ID, roles[0].ID); err == nil {
		t.Fatal("expected error deleting system role")
	}
	if err := roleSvc.DeleteGroupRole(ctx, owner.AccessToken, group.ID, dup.PublicID); err != nil {
		t.Fatalf("DeleteGroupRole: %v", err)
	}
}
