import {
	CreateMessageBodySchema,
	CreateMessageResponseSchema,
	MessagesListSchema,
	type CreateMessageBody,
	type CreateMessageResponse,
	type MessagesList,
} from "@wichat/contracts";
import { api } from "@/integrations/api-client";

export async function listChannelMessages(
	groupId: string,
	channelId: string,
): Promise<MessagesList> {
	return api.get(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/messages`,
		MessagesListSchema,
	);
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
