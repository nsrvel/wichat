import type { GroupMember } from "@wichat/contracts/group/member";
import type { RecentPersonHit } from "@/features/search/lib/recent-interaction-people";

const MAX_HITS = 8;

function memberMatchesQuery(member: GroupMember, q: string): boolean {
	const needle = q.toLowerCase();
	const display = (member.display_name ?? "").toLowerCase();
	const username = (member.username ?? "").toLowerCase();
	return display.includes(needle) || username.includes(needle);
}

export function filterGroupMembersForSearch(
	members: GroupMember[] | undefined,
	query: string,
	excludeUserId?: string,
): RecentPersonHit[] {
	const q = query.trim();
	if (q.length < 2 || !members?.length) {
		return [];
	}

	const hits: RecentPersonHit[] = [];
	for (const member of members) {
		if (excludeUserId && member.user_id === excludeUserId) {
			continue;
		}
		if (!memberMatchesQuery(member, q)) {
			continue;
		}
		hits.push({
			userId: member.user_id,
			displayName: member.display_name || member.username || "",
			username: member.username,
			avatarUrl: member.avatar_url ?? "",
			is_online: member.is_online,
			presence_status: member.presence_status,
		});
		if (hits.length >= MAX_HITS) {
			break;
		}
	}
	return hits;
}
