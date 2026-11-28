import { queryOptions } from "@tanstack/react-query";
import {
	countGroupMentionsUnread,
	listGroupMentions,
	type ListGroupMentionsParams,
} from "@/services/mentions/mentions.service";

export const Q_GROUP_MENTIONS = "mentions.group";
export const Q_GROUP_MENTIONS_UNREAD = "mentions.group.unread_count";

export function groupMentionsQueryOpts(groupId: string, params: ListGroupMentionsParams = {}) {
	return queryOptions({
		queryKey: [Q_GROUP_MENTIONS, groupId, params.kind ?? "all", params.cursor ?? ""],
		queryFn: () => listGroupMentions(groupId, params),
		staleTime: 15_000,
		enabled: Boolean(groupId),
	});
}

export function groupMentionsUnreadCountQueryOpts(groupId: string) {
	return queryOptions({
		queryKey: [Q_GROUP_MENTIONS_UNREAD, groupId],
		queryFn: () => countGroupMentionsUnread(groupId),
		staleTime: 15_000,
		enabled: Boolean(groupId),
	});
}
