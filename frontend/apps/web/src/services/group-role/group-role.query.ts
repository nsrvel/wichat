import { queryOptions } from "@tanstack/react-query";
import { groupRoleKeys } from "@/services/group-role/group-role.key";
import { listGroupRoleMembers, listGroupRoles } from "@/services/group-role/group-role.service";

export function groupRolesQueryOpts(groupId: string) {
	return queryOptions({
		queryKey: groupRoleKeys.list(groupId),
		queryFn: () => listGroupRoles(groupId),
		staleTime: 30_000,
		enabled: Boolean(groupId),
	});
}

export function groupRoleMembersQueryOpts(groupId: string, roleId: string) {
	return queryOptions({
		queryKey: groupRoleKeys.members(groupId, roleId),
		queryFn: () => listGroupRoleMembers(groupId, roleId),
		staleTime: 30_000,
		enabled: Boolean(groupId && roleId),
	});
}
