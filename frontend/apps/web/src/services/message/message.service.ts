import {
	CreateMessageBodySchema,
	CreateMessageResponseSchema,
	MessagesListSchema,
	ReactionsResponseSchema,
	ToggleReactionBodySchema,
	UpdateMessageBodySchema,
	UploadAttachmentResponseSchema,
	type CreateMessageBody,
	type CreateMessageResponse,
	type MessageAttachment,
	type MessagesList,
	type ReactionsResponse,
	type ToggleReactionBody,
	type UpdateMessageBody,
} from "@wichat/contracts/chat/message";
import { OkSchema } from "@wichat/contracts/common";
import { api } from "@/integrations/api-client";

export const CHANNEL_MESSAGES_PAGE_SIZE = 50;

export type ListChannelMessagesParams = {
	limit?: number;
	beforeCreatedAt?: string;
	beforeId?: string;
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
	if (params?.beforeId) {
		search.set("before_id", params.beforeId);
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

export async function updateChannelMessage(
	groupId: string,
	channelId: string,
	messageId: string,
	body: UpdateMessageBody,
): Promise<CreateMessageResponse> {
	UpdateMessageBodySchema.parse(body);
	return api.patch(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/messages/${encodeURIComponent(messageId)}`,
		CreateMessageResponseSchema,
		body,
	);
}

export async function deleteChannelMessage(groupId: string, channelId: string, messageId: string) {
	return api.delete(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/messages/${encodeURIComponent(messageId)}`,
		OkSchema,
	);
}

export async function toggleMessageReaction(
	groupId: string,
	channelId: string,
	messageId: string,
	body: ToggleReactionBody,
): Promise<ReactionsResponse> {
	ToggleReactionBodySchema.parse(body);
	return api.post(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/messages/${encodeURIComponent(messageId)}/reactions`,
		ReactionsResponseSchema,
		body,
	);
}

export async function listPinnedChannelMessages(
	groupId: string,
	channelId: string,
): Promise<MessagesList> {
	return api.get(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/pins`,
		MessagesListSchema,
	);
}

export async function pinChannelMessage(
	groupId: string,
	channelId: string,
	messageId: string,
): Promise<CreateMessageResponse> {
	return api.put(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/pins/${encodeURIComponent(messageId)}`,
		CreateMessageResponseSchema,
		{},
	);
}

export async function unpinChannelMessage(
	groupId: string,
	channelId: string,
	messageId: string,
): Promise<CreateMessageResponse> {
	return api.delete(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/pins/${encodeURIComponent(messageId)}`,
		CreateMessageResponseSchema,
	);
}

export async function uploadChannelAttachment(
	groupId: string,
	channelId: string,
	file: File,
): Promise<MessageAttachment> {
	const form = new FormData();
	form.append("file", file);
	const res = await api.postForm(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/attachments`,
		UploadAttachmentResponseSchema,
		form,
	);
	return res.attachment;
}
