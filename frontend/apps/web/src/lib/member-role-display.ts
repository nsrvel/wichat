import type { GroupMember, GroupMemberRoleSummary } from "@wichat/contracts/group/member";

/**
 * Default neutral role colors that should not override the member's default
 * foreground text color.
 */
const DEFAULT_ROLE_COLORS = new Set(["", "#64748b", "#99aab5"]);

/**
 * Returns true if the given role color is a customized non-default color.
 */
export function hasCustomRoleColor(color: string | undefined): boolean {
	if (!color) {
		return false;
	}
	const trimmed = color.trim().toLowerCase();
	return Boolean(trimmed) && !DEFAULT_ROLE_COLORS.has(trimmed);
}

/**
 * Resolves the effective font color for a group member's display name based on
 * their ordered roles (highest priority first) and the group's `hide_roles`
 * setting.
 *
 * When `hideRoles` is true, all role display customizations are suppressed and
 * `undefined` is returned so the UI falls back to the default text color.
 */
export function resolveMemberNameColor(
	roles: readonly GroupMemberRoleSummary[] | undefined,
	hideRoles?: boolean,
): string | undefined {
	if (hideRoles || !roles || roles.length === 0) {
		return undefined;
	}
	for (const role of roles) {
		const color = role.color.trim();
		if (hasCustomRoleColor(color)) {
			return color;
		}
	}
	return undefined;
}

/**
 * Resolves the highest-priority role with `hoist === true` for a member.
 * Returns `null` when `hideRoles` is enabled or no hoisted role is assigned.
 */
export function resolveMemberHoistedRole(
	roles: readonly GroupMemberRoleSummary[] | undefined,
	hideRoles?: boolean,
): GroupMemberRoleSummary | null {
	if (hideRoles || !roles || roles.length === 0) {
		return null;
	}
	for (const role of roles) {
		if (role.hoist) {
			return role;
		}
	}
	return null;
}

export type MemberPanelGroup = {
	key: string;
	label: string;
	kind: "hoisted" | "online" | "offline";
	position?: number;
	members: GroupMember[];
};

/**
 * Groups members for the workspace right-hand MemberPanel:
 * - When `hideRoles` is false: online members with a hoisted role are grouped
 *   under that role (ordered by `position ASC`), followed by remaining online
 *   members, then offline members.
 * - When `hideRoles` is true: all role display customizations are suppressed,
 *   producing only `Online` and `Offline` sections.
 */
export function groupMembersForPanel(
	members: readonly GroupMember[],
	hideRoles: boolean,
	labels: { online: string; offline: string },
): MemberPanelGroup[] {
	const hoistedMap = new Map<string, { role: GroupMemberRoleSummary; members: GroupMember[] }>();
	const online: GroupMember[] = [];
	const offline: GroupMember[] = [];

	for (const member of members) {
		if (!member.is_online) {
			offline.push(member);
			continue;
		}
		const hoisted = resolveMemberHoistedRole(member.roles, hideRoles);
		if (hoisted) {
			const existing = hoistedMap.get(hoisted.id);
			if (existing) {
				existing.members.push(member);
			} else {
				hoistedMap.set(hoisted.id, { role: hoisted, members: [member] });
			}
			continue;
		}
		online.push(member);
	}

	const groups: MemberPanelGroup[] = [];
	const sortedHoisted = Array.from(hoistedMap.values()).sort(
		(a, b) => a.role.position - b.role.position,
	);
	for (const entry of sortedHoisted) {
		groups.push({
			key: `role:${entry.role.id}`,
			label: entry.role.name,
			kind: "hoisted",
			position: entry.role.position,
			members: entry.members,
		});
	}
	if (online.length > 0) {
		groups.push({
			key: "online",
			label: labels.online,
			kind: "online",
			members: online,
		});
	}
	if (offline.length > 0) {
		groups.push({
			key: "offline",
			label: labels.offline,
			kind: "offline",
			members: offline,
		});
	}
	return groups;
}
