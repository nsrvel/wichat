export const Q_GROUP_ROLES = "group-roles";
export const Q_GROUP_ROLE_MEMBERS = "group-role-members";

export const groupRoleKeys = {
	list: (groupId: string) => [Q_GROUP_ROLES, groupId] as const,
	members: (groupId: string, roleId: string) => [Q_GROUP_ROLE_MEMBERS, groupId, roleId] as const,
};
