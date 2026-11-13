import {
	CreateMessageBodySchema,
	CreateMessageResponseSchema,
	MessagesListSchema,
	type CreateMessageBody,
	type CreateMessageResponse,
	type MessagesList,
} from "@wichat/contracts";
import { api } from "@/integrations/api-client";

export const CHANNEL_MESSAGES_PAGE_SIZE = 50;

export type ListChannelMessagesParams = {
	limit?: number;
	beforeCreatedAt?: string;
};

export async function listChannelMessages(
	groupId: string,
	channelId: string,
	params?: ListChannelMessagesParams,
): Promise<MessagesList> {
	const search = new URLSearchParams();
	search.set("limit", String(params?.limit ?? CHANNEL_MESSAGES_PAGE_SIZE));
	if (params?.beforeCreatedAt) {
		search.set("before_created_at", params.beforeCreatedAt);
	}
	const query = search.toString();
	const path = `/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/messages${query ? `?${query}` : ""}`;
	return api.get(path, MessagesListSchema);
}

export async function createChannelMessage(
	groupId: string,
	channelId: string,
	body: CreateMessageBody,
): Promise<CreateMessageResponse> {
	CreateMessageBodySchema.parse(body);
	return api.post(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/messages`,
		CreateMessageResponseSchema,
		body,
	);
}
