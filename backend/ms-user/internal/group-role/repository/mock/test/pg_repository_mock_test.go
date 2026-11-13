package test

import (
	"context"
	"testing"
	"time"

	"github.com/uptrace/bun"
	grouprolemock "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestPgRepositoryMock(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	m := grouprolemock.NewPgRepositoryMock(st)

	groupID := "grp-1"
	ownerID := "usr-owner"
	memberID := "usr-member"
	now := time.Now().UTC()

	st.Members = append(st.Members,
		models.GroupMember{GroupID: groupID, UserID: ownerID, Role: constants.GroupRoleOwner, CreatedAt: now},
		models.GroupMember{GroupID: groupID, UserID: memberID, Role: constants.GroupRoleMember, CreatedAt: now},
	)

	if err := m.SeedDefaultGroupRolesTx(ctx, bun.Tx{}, groupID, ownerID, now); err != nil {
		t.Fatalf("SeedDefaultGroupRolesTx failed: %v", err)
	}

	roles, err := m.ListRolesForGroup(ctx, groupID)
	if err != nil || len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d, err: %v", len(roles), err)
	}

	perms, err := m.GetEffectivePermissions(ctx, groupID, ownerID)
	if err != nil || len(perms) != len(constants.AllKnownPermissions) {
		t.Fatalf("expected owner to have all permissions, got %d", len(perms))
	}

	role, err := m.FindRoleByID(ctx, groupID, roles[1].ID)
	if err != nil {
		t.Fatalf("FindRoleByID failed: %v", err)
	}

	newName := "Moderator"
	patch := models.GroupRolePatch{
		Name:        &newName,
		Permissions: []string{constants.PermManageRoles, constants.PermKickMembers},
	}
	updated, err := m.UpdateRole(ctx, groupID, role.ID, patch, now)
	if err != nil || updated.Name != "Moderator" {
		t.Fatalf("UpdateRole failed: %v", err)
	}

	if err := m.AddMemberRoles(ctx, groupID, role.ID, []string{memberID}); err != nil {
		t.Fatalf("AddMemberRoles failed: %v", err)
	}

	memberPerms, err := m.GetEffectivePermissions(ctx, groupID, memberID)
	if err != nil {
		t.Fatalf("GetEffectivePermissions failed: %v", err)
	}
	hasKick := false
	for _, p := range memberPerms {
		if p == constants.PermKickMembers {
			hasKick = true
			break
		}
	}
	if !hasKick {
		t.Fatalf("expected member to have kick permission after role assignment")
	}

	roleMap, err := m.ListRolesForMembers(ctx, groupID, []string{memberID})
	if err != nil || len(roleMap[memberID]) == 0 {
		t.Fatalf("ListRolesForMembers failed: %v", err)
	}

	if err := m.RemoveMemberRole(ctx, groupID, role.ID, memberID); err != nil {
		t.Fatalf("RemoveMemberRole failed: %v", err)
	}

	_, err = m.FindRoleByID(ctx, groupID, "non-existent")
	if err == nil {
		t.Fatalf("expected not found for non-existent role")
	}
}
