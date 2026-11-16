import type { GroupMember, GroupMemberRoleSummary } from "@wichat/contracts/group/member";
import type { GroupRole } from "@wichat/contracts/group/role";
import { hasGroupPermission, type GroupPermissionSubject } from "@/lib/group-permissions";

export type MemberModerationPolicy = {
	isSelf: boolean;
	targetIsOwner: boolean;
	canTimeout: boolean;
	canKick: boolean;
	canBan: boolean;
	canManageRoles: boolean;
	assignableRoles: GroupRole[];
	canBulkSelect: boolean;
};

/**
 * Returns the lowest `position` number (highest rank) among a member's non-default roles,
 * or `Number.POSITIVE_INFINITY` if the member only holds `@everyone`.
 */
export function resolveBestCustomRolePosition(
	roles: readonly Pick<GroupMemberRoleSummary, "slug" | "position">[] | undefined,
): number {
	if (!roles || roles.length === 0) {
		return Number.POSITIVE_INFINITY;
	}
	let best = Number.POSITIVE_INFINITY;
	for (const role of roles) {
		if (role.slug === "everyone") {
			continue;
		}
		if (role.position < best) {
			best = role.position;
		}
	}
	return best;
}

/**
 * Resolves the caller's best custom role position from the visible member roster
 * (falling back to `Number.POSITIVE_INFINITY` when not found).
 */
export function resolveCallerBestRolePosition(
	actorUserId: string | undefined,
	members: readonly GroupMember[],
): number {
	if (!actorUserId) {
		return Number.POSITIVE_INFINITY;
	}
	const self = members.find((m) => m.user_id === actorUserId);
	return resolveBestCustomRolePosition(self?.roles);
}

/**
 * Computes what moderation & role-assignment actions the current user may perform
 * on a specific target member, enforcing self-guards, owner-guards, permissions,
 * and Discord-style role hierarchy (`actorBestPosition < targetBestPosition`).
 */
export function getMemberModerationPolicy(params: {
	actorUserId?: string;
	actorGroup: GroupPermissionSubject | null | undefined;
	actorBestRolePosition: number;
	target: GroupMember;
	allRoles: readonly GroupRole[];
}): MemberModerationPolicy {
	const { actorUserId, actorGroup, actorBestRolePosition, target, allRoles } = params;
	const isSelf = Boolean(actorUserId && target.user_id === actorUserId);
	const targetIsOwner = target.role === "owner";
	const actorIsOwner = actorGroup?.role === "owner";

	const targetBestPosition = resolveBestCustomRolePosition(target.roles);
	const outranksTarget =
		actorIsOwner ||
		(!targetIsOwner &&
			(targetBestPosition === Number.POSITIVE_INFINITY ||
				actorBestRolePosition < targetBestPosition));

	const hasTimeoutPerm = hasGroupPermission(actorGroup, "timeout_members");
	const hasKickPerm = hasGroupPermission(actorGroup, "kick_members");
	const hasBanPerm = hasGroupPermission(actorGroup, "ban_members");
	const hasManageRolesPerm = hasGroupPermission(actorGroup, "manage_roles");

	const canTimeout = !isSelf && !targetIsOwner && hasTimeoutPerm && outranksTarget;
	const canKick = !isSelf && !targetIsOwner && hasKickPerm && outranksTarget;
	const canBan = !isSelf && !targetIsOwner && hasBanPerm && outranksTarget;
	const canManageRoles = hasManageRolesPerm && (actorIsOwner || !targetIsOwner);

	const customRoles = allRoles
		.filter((r) => !r.is_default && r.slug !== "everyone")
		.sort((a, b) => a.position - b.position);

	const assignableRoles = canManageRoles
		? customRoles.filter((r) => actorIsOwner || r.position > actorBestRolePosition)
		: [];

	const canBulkSelect = !isSelf && !targetIsOwner && (canKick || canBan);

	return {
		isSelf,
		targetIsOwner,
		canTimeout,
		canKick,
		canBan,
		canManageRoles,
		assignableRoles,
		canBulkSelect,
	};
}
