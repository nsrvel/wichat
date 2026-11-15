export type GroupPermissionSubject = {
	role?: string;
	effective_permissions?: readonly string[];
};

/**
 * Checks whether the current user holds the given permission in a group.
 * Group owners and holders of `administrator` implicitly satisfy every permission.
 */
export function hasGroupPermission(
	group: GroupPermissionSubject | null | undefined,
	permission: string,
): boolean {
	if (!group) {
		return false;
	}
	if (group.role === "owner") {
		return true;
	}
	const perms = group.effective_permissions ?? [];
	return perms.includes("administrator") || perms.includes(permission);
}

/**
 * Checks whether the current user holds at least one of the listed permissions.
 */
export function hasAnyGroupPermission(
	group: GroupPermissionSubject | null | undefined,
	permissions: readonly string[],
): boolean {
	return permissions.some((permission) => hasGroupPermission(group, permission));
}
