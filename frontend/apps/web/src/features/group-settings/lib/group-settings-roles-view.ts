/** Built-in membership roles (custom roles API deferred). */
export type GroupSettingsBuiltinRoleId = "owner" | "member";

export type GroupSettingsRoleRow = {
	id: string;
	/** Stable key for built-ins; display name for custom roles later. */
	name: string;
	memberCount: number;
	isSystem: boolean;
};

export function builtinRoleRows(memberTotal: number): GroupSettingsRoleRow[] {
	const members = Math.max(0, memberTotal - 1);
	return [
		{ id: "owner", name: "owner", memberCount: memberTotal > 0 ? 1 : 0, isSystem: true },
		{ id: "member", name: "member", memberCount: members, isSystem: true },
	];
}

export function matchesRoleSearch(
	row: GroupSettingsRoleRow,
	query: string,
	roleLabel: string,
): boolean {
	const q = query.trim().toLowerCase();
	if (!q) {
		return true;
	}
	return (
		row.name.toLowerCase().includes(q) ||
		row.id.toLowerCase().includes(q) ||
		roleLabel.toLowerCase().includes(q)
	);
}
