package access

import "github.com/wichat/wichat/backend/ms-user/pkg/constants"

// HasPermission checks if the effective permissions list contains the requested permission.
// Administrator permission or group owner grants all permissions.
func HasPermission(isOwner bool, permissions []string, targetPerm string) bool {
	if isOwner {
		return true
	}
	for _, p := range permissions {
		if p == constants.PermAdministrator || p == targetPerm {
			return true
		}
	}
	return false
}

// ComputeEffectivePermissions calculates the set of permissions a user has in a group.
// It combines default (@everyone) permissions with any role-assigned permissions.
// If the user is the group owner or has administrator, all known permissions are returned.
func ComputeEffectivePermissions(isOwner bool, everyonePerms []string, rolePerms [][]string) []string {
	if isOwner {
		return append([]string(nil), constants.AllKnownPermissions...)
	}

	set := make(map[string]struct{}, len(everyonePerms))
	for _, p := range everyonePerms {
		set[p] = struct{}{}
	}
	for _, rPerms := range rolePerms {
		for _, p := range rPerms {
			set[p] = struct{}{}
		}
	}

	if _, hasAdmin := set[constants.PermAdministrator]; hasAdmin {
		return append([]string(nil), constants.AllKnownPermissions...)
	}

	out := make([]string, 0, len(set))
	for _, p := range constants.AllKnownPermissions {
		if _, ok := set[p]; ok {
			out = append(out, p)
		}
	}
	return out
}
