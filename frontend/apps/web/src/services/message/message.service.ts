import {
	CreateMessageBodySchema,
	CreateMessageResponseSchema,
	MessageChannelSeqSchema,
	MessagesListSchema,
	ReactionsResponseSchema,
	ToggleReactionBodySchema,
	UpdateMessageBodySchema,
	UploadAttachmentResponseSchema,
	type CreateMessageBody,
	type CreateMessageResponse,
	type MessageAttachment,
	type MessageChannelSeq,
	type MessagesList,
	type ReactionsResponse,
	type ToggleReactionBody,
	type UpdateMessageBody,
} from "@wichat/contracts/chat/message";
import { OkSchema } from "@wichat/contracts/common";
import { api } from "@/integrations/api-client";

export const CHANNEL_MESSAGES_PAGE_SIZE = 50;
const JUMP_CONTEXT_HALF = 60;

export type ListChannelMessagesParams = {
	limit?: number;
	beforeChannelSeq?: number;
	afterChannelSeq?: number;
	signal?: AbortSignal;
};

export async function listChannelMessages(
	groupId: string,
	channelId: string,
	params?: ListChannelMessagesParams,
): Promise<MessagesList> {
	const search = new URLSearchParams();
	search.set("limit", String(params?.limit ?? CHANNEL_MESSAGES_PAGE_SIZE));
	if (params?.beforeChannelSeq != null && params.beforeChannelSeq > 0) {
		search.set("before_channel_seq", String(params.beforeChannelSeq));
	}
	if (params?.afterChannelSeq != null && params.afterChannelSeq > 0) {
		search.set("after_channel_seq", String(params.afterChannelSeq));
	}
	const query = search.toString();
	const path = `/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/messages${query ? `?${query}` : ""}`;
	return api.get(path, MessagesListSchema, { signal: params?.signal });
}

export async function resolveMessageChannelSeq(
	groupId: string,
	channelId: string,
	messageId: string,
	signal?: AbortSignal,
): Promise<MessageChannelSeq> {
	return api.get(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/messages/${encodeURIComponent(messageId)}/seq`,
		MessageChannelSeqSchema,
		{ signal },
	);
}

export type ChannelMessagesContextWindow = {
	messages: MessagesList["messages"];
	hasOlder: boolean;
	hasNewer: boolean;
};

/** Fetches a bounded window around channel_seq (no walk from latest). */
export async function listChannelMessagesAroundSeq(
	groupId: string,
	channelId: string,
	centerSeq: number,
	opts?: { half?: number; signal?: AbortSignal },
): Promise<ChannelMessagesContextWindow> {
	if (centerSeq <= 0) {
		return { messages: [], hasOlder: false, hasNewer: false };
	}
	const half = opts?.half ?? JUMP_CONTEXT_HALF;
	const [older, newer] = await Promise.all([
		listChannelMessages(groupId, channelId, {
			limit: half,
			beforeChannelSeq: centerSeq,
			signal: opts?.signal,
		}),
		listChannelMessages(groupId, channelId, {
			limit: half,
			afterChannelSeq: Math.max(0, centerSeq - 1),
			signal: opts?.signal,
		}),
	]);

	const byId = new Map<string, MessagesList["messages"][number]>();
	for (const msg of older.messages) {
		byId.set(msg.id, msg);
	}
	for (const msg of newer.messages) {
		byId.set(msg.id, msg);
	}
	const messages = [...byId.values()].sort(
		(a, b) => (a.channel_seq ?? 0) - (b.channel_seq ?? 0) || a.id.localeCompare(b.id),
	);

	return {
		messages,
		hasOlder: older.has_more !== false && older.messages.length > 0,
		hasNewer: newer.has_more !== false && newer.messages.length > 0,
	};
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
