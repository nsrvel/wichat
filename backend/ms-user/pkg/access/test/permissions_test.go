package access_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/access"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestHasPermission(t *testing.T) {
	// Owner bypasses check
	if !access.HasPermission(true, nil, constants.PermManageGroup) {
		t.Fatal("expected owner to have all permissions")
	}

	// Admin role bypasses check
	if !access.HasPermission(false, []string{constants.PermAdministrator}, constants.PermKickMembers) {
		t.Fatal("expected administrator to grant kick_members")
	}

	// Explicit permission
	if !access.HasPermission(false, []string{constants.PermSendMessages}, constants.PermSendMessages) {
		t.Fatal("expected send_messages to be allowed")
	}

	// Missing permission
	if access.HasPermission(false, []string{constants.PermSendMessages}, constants.PermManageRoles) {
		t.Fatal("expected manage_roles to be denied")
	}
}

func TestComputeEffectivePermissions(t *testing.T) {
	// Owner gets all
	ownerPerms := access.ComputeEffectivePermissions(true, nil, nil)
	if len(ownerPerms) != len(constants.AllKnownPermissions) {
		t.Fatalf("expected all permissions for owner, got %d", len(ownerPerms))
	}

	// Normal user unions everyone and assigned roles
	everyone := []string{constants.PermViewChannels, constants.PermSendMessages}
	role1 := []string{constants.PermManageGroup}
	role2 := []string{constants.PermCreateInvite}

	effective := access.ComputeEffectivePermissions(false, everyone, [][]string{role1, role2})
	if len(effective) != 4 {
		t.Fatalf("expected 4 effective permissions, got %d", len(effective))
	}

	// Admin in roles grants all
	roleAdmin := []string{constants.PermAdministrator}
	effectiveAdmin := access.ComputeEffectivePermissions(false, everyone, [][]string{roleAdmin})
	if len(effectiveAdmin) != len(constants.AllKnownPermissions) {
		t.Fatalf("expected all permissions for admin role, got %d", len(effectiveAdmin))
	}
}
