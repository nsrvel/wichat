import {
	GroupMentionsListSchema,
	MarkMentionsReadBodySchema,
	MentionsUnreadCountSchema,
	type GroupMentionsList,
	type MarkMentionsReadBody,
	type MentionsUnreadCount,
} from "@wichat/contracts/chat/mentions";
import { OkSchema } from "@wichat/contracts/common";
import { api } from "@/integrations/api-client";

export type ListGroupMentionsParams = {
	limit?: number;
	cursor?: string;
	kind?: "all" | "channels" | "chats" | "unread";
};

export async function listGroupMentions(
	groupId: string,
	params: ListGroupMentionsParams = {},
): Promise<GroupMentionsList> {
	const search = new URLSearchParams();
	if (params.limit) {
		search.set("limit", String(params.limit));
	}
	if (params.cursor?.trim()) {
		search.set("cursor", params.cursor.trim());
	}
	if (params.kind && params.kind !== "all") {
		search.set("kind", params.kind);
	}
	const qs = search.toString();
	const path = `/groups/${encodeURIComponent(groupId)}/mentions${qs ? `?${qs}` : ""}`;
	return api.get(path, GroupMentionsListSchema);
}

export async function countGroupMentionsUnread(groupId: string): Promise<MentionsUnreadCount> {
	return api.get(
		`/groups/${encodeURIComponent(groupId)}/mentions/unread-count`,
		MentionsUnreadCountSchema,
	);
}

export async function markGroupMentionsRead(
	groupId: string,
	body: MarkMentionsReadBody,
): Promise<void> {
	MarkMentionsReadBodySchema.parse(body);
	await api.post(`/groups/${encodeURIComponent(groupId)}/mentions/read`, OkSchema, body);
}
