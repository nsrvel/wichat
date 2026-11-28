import type { GroupRole } from "@wichat/contracts/group/role";

const PERM_MENTION_EVERYONE = "mention_everyone";

export function canMentionEveryone(callerPermissions: string[] | undefined): boolean {
	const perms = callerPermissions ?? [];
	return perms.includes("administrator") || perms.includes(PERM_MENTION_EVERYONE);
}

export function isRoleVisibleInMentionPicker(
	role: GroupRole,
	callerPermissions: string[] | undefined,
): boolean {
	if (role.is_system && role.slug === "everyone") {
		return canMentionEveryone(callerPermissions);
	}
	return role.mentionable || canMentionEveryone(callerPermissions);
}

export function mentionPickerRoles(
	roles: GroupRole[],
	callerPermissions: string[] | undefined,
	query: string,
): GroupRole[] {
	const needle = query.trim().toLowerCase();
	const visible = roles.filter((r) => isRoleVisibleInMentionPicker(r, callerPermissions));
	const filtered = !needle ? visible : visible.filter((r) => r.name.toLowerCase().includes(needle));
	return filtered.sort((a, b) => a.position - b.position).slice(0, 12);
}
