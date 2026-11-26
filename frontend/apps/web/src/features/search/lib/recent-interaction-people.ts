import type { InfiniteData } from "@tanstack/react-query";
import type { GroupMember } from "@wichat/contracts/group/member";
import type { MessagesList } from "@wichat/contracts/chat/message";
import { flattenInfiniteChannelMessages } from "@/features/chat/lib/channel-message";
import {
	memberPresenceDot,
	type AvatarPresenceDot,
} from "@/features/presence/presence-dot";

export type RecentPersonHit = {
	userId: string;
	displayName: string;
	username: string;
	avatarUrl: string;
	is_online?: GroupMember["is_online"];
	presence_status?: GroupMember["presence_status"];
};

export function recentPersonPresenceDot(
	person: Pick<RecentPersonHit, "is_online" | "presence_status">,
): AvatarPresenceDot | undefined {
	if (person.is_online === undefined && person.presence_status === undefined) {
		return undefined;
	}
	return memberPresenceDot({
		is_online: person.is_online ?? false,
		presence_status: person.presence_status ?? "offline",
	});
}

function personHitFromMember(member: GroupMember): RecentPersonHit {
	return {
		userId: member.user_id,
		displayName: member.display_name || member.username,
		username: member.username,
		avatarUrl: member.avatar_url || "",
		is_online: member.is_online,
		presence_status: member.presence_status,
	};
}

function memberFieldsForUser(
	member: GroupMember | undefined,
): Pick<RecentPersonHit, "is_online" | "presence_status"> {
	if (!member) {
		return {};
	}
	return {
		is_online: member.is_online,
		presence_status: member.presence_status,
	};
}

const MAX_PEOPLE = 8;

export function recentPeopleFromChannelCache(
	data: InfiniteData<MessagesList> | undefined,
	memberById: Map<string, GroupMember>,
	excludeUserId?: string,
): RecentPersonHit[] {
	const messages = flattenInfiniteChannelMessages(data);
	const order: string[] = [];
	const seen = new Set<string>();

	for (let i = messages.length - 1; i >= 0; i -= 1) {
		const authorId = messages[i]?.author_id;
		if (!authorId || authorId === excludeUserId || seen.has(authorId)) {
			continue;
		}
		seen.add(authorId);
		order.push(authorId);
		if (order.length >= MAX_PEOPLE) {
			break;
		}
	}

	const hits: RecentPersonHit[] = [];
	for (const userId of order) {
		const member = memberById.get(userId);
		const msg = messages.find((m) => m.author_id === userId);
		const displayName =
			member?.display_name || msg?.author_display_name || member?.username || userId;
		const username = member?.username || "";
		hits.push({
			userId,
			displayName,
			username,
			avatarUrl: member?.avatar_url || msg?.author_avatar_url || "",
			...memberFieldsForUser(member),
		});
	}

	if (hits.length >= MAX_PEOPLE) {
		return hits;
	}

	for (const member of memberById.values()) {
		if (hits.length >= MAX_PEOPLE) {
			break;
		}
		if (member.user_id === excludeUserId || seen.has(member.user_id)) {
			continue;
		}
		seen.add(member.user_id);
		hits.push(personHitFromMember(member));
	}

	return hits.slice(0, MAX_PEOPLE);
}
