import { GroupMemberSchema } from "@wichat/contracts";
import type { z } from "zod";
import { parseUtcIso } from "@/features/group-settings/lib/group-settings-members-format";

type GroupMember = z.infer<typeof GroupMemberSchema>;

/** How a user entered the group — API will supply this later. */
export type MemberJoinMethod = "invite" | "invite_link" | "open" | "unknown";

export type MembersSortKey =
	| "name_asc"
	| "name_desc"
	| "member_since_asc"
	| "member_since_desc"
	| "last_updated_asc"
	| "last_updated_desc"
	| "role_asc";

export type MembersPruneKey = "all" | "inactive_7d" | "inactive_30d" | "inactive_90d";

/** Inactivity window in the prune-members dialog (kick flow). */
export type MembersPruneInactiveKey = "inactive_7d" | "inactive_30d";

export type PruneDialogRoleKey = "owner" | "member";

export type GroupSettingsMemberRow = {
	member: GroupMember;
	/** Group membership start (`joined_at`). */
	memberSinceAt: string;
	/** WiChat account age — not on members API yet. */
	joinedWiChatAt: string | null;
	joinMethod: MemberJoinMethod;
	/** Primary + future extra roles (today: one role from API). */
	roles: string[];
	/**
	 * Activity for sort/prune — ponytail: `joined_at` until presence/last_seen is on ListGroupMembers.
	 */
	lastUpdatedAt: string;
};

export function toMemberRow(member: GroupMember): GroupSettingsMemberRow {
	return {
		member,
		memberSinceAt: member.joined_at,
		joinedWiChatAt: null,
		joinMethod: "unknown",
		roles: [member.role],
		lastUpdatedAt: member.joined_at,
	};
}

export function matchesMemberSearch(row: GroupSettingsMemberRow, query: string): boolean {
	const q = query.trim().toLowerCase();
	if (!q) {
		return true;
	}
	const { display_name, username } = row.member;
	return (
		display_name.trim().toLowerCase().includes(q) ||
		(username ?? "").trim().toLowerCase().includes(q)
	);
}

export function passesPrune(
	row: GroupSettingsMemberRow,
	prune: MembersPruneKey,
	now = Date.now(),
): boolean {
	if (prune === "all") {
		return true;
	}
	const activeAt = parseUtcIso(row.lastUpdatedAt);
	if (!activeAt) {
		return false;
	}
	const days = prune === "inactive_7d" ? 7 : prune === "inactive_30d" ? 30 : 90;
	const cutoff = now - days * 86400 * 1000;
	return activeAt.getTime() < cutoff;
}

function compareIso(a: string, b: string): number {
	const da = parseUtcIso(a)?.getTime() ?? 0;
	const db = parseUtcIso(b)?.getTime() ?? 0;
	return da - db;
}

export function sortMemberRows(
	rows: GroupSettingsMemberRow[],
	sort: MembersSortKey,
): GroupSettingsMemberRow[] {
	const next = [...rows];
	next.sort((a, b) => {
		switch (sort) {
			case "name_asc":
				return a.member.display_name.localeCompare(b.member.display_name, undefined, {
					sensitivity: "base",
				});
			case "name_desc":
				return b.member.display_name.localeCompare(a.member.display_name, undefined, {
					sensitivity: "base",
				});
			case "member_since_asc":
				return compareIso(a.memberSinceAt, b.memberSinceAt);
			case "member_since_desc":
				return compareIso(b.memberSinceAt, a.memberSinceAt);
			case "last_updated_asc":
				return compareIso(a.lastUpdatedAt, b.lastUpdatedAt);
			case "last_updated_desc":
				return compareIso(b.lastUpdatedAt, a.lastUpdatedAt);
			case "role_asc":
				return a.member.role.localeCompare(b.member.role);
			default:
				return 0;
		}
	});
	return next;
}

export function applyMemberView(
	rows: GroupSettingsMemberRow[],
	opts: { search: string; sort: MembersSortKey },
): GroupSettingsMemberRow[] {
	const filtered = rows.filter((r) => matchesMemberSearch(r, opts.search));
	return sortMemberRows(filtered, opts.sort);
}

function pruneInactiveDays(key: MembersPruneInactiveKey): number {
	return key === "inactive_7d" ? 7 : 30;
}

/** Members eligible for prune preview — ponytail: uses `joined_at` as last seen until presence API ships. */
export function isPruneKickCandidate(
	row: GroupSettingsMemberRow,
	inactive: MembersPruneInactiveKey,
	alsoRoles: readonly PruneDialogRoleKey[],
	now = Date.now(),
): boolean {
	if (!passesPrune(row, inactive, now)) {
		return false;
	}
	const role = row.member.role;
	if (role === "owner") {
		return alsoRoles.includes("owner");
	}
	if (role === "member") {
		return true;
	}
	return alsoRoles.includes(role as PruneDialogRoleKey);
}

export function countPruneKickCandidates(
	rows: GroupSettingsMemberRow[],
	inactive: MembersPruneInactiveKey,
	alsoRoles: readonly PruneDialogRoleKey[],
	now = Date.now(),
): number {
	return rows.filter((r) => isPruneKickCandidate(r, inactive, alsoRoles, now)).length;
}

export function pruneInactiveDayCount(inactive: MembersPruneInactiveKey): number {
	return pruneInactiveDays(inactive);
}
