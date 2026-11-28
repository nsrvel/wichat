export type MentionMemberRow = {
	user_id: string;
	display_name: string;
	username: string;
	avatar_url?: string | null;
};

export type MentionSuggestMemberItem = {
	kind: "user";
	id: string;
	label: string;
	username: string;
	avatarUrl: string;
};

export function filterMentionSuggestMembers(
	members: MentionMemberRow[],
	query: string,
	selfUserId?: string,
): MentionSuggestMemberItem[] {
	const needle = query.trim().toLowerCase();
	const self = selfUserId?.trim() ?? "";
	const filtered = members.filter((m) => {
		if (self !== "" && m.user_id === self) {
			return false;
		}
		if (!needle) {
			return true;
		}
		const name = m.display_name?.toLowerCase() ?? "";
		const username = m.username?.toLowerCase() ?? "";
		return name.includes(needle) || username.includes(needle);
	});
	const rows = filtered.slice(0, 12);
	return rows.map((m) => ({
		kind: "user",
		id: m.user_id,
		label: m.display_name || m.username,
		username: m.username,
		avatarUrl: m.avatar_url ?? "",
	}));
}
