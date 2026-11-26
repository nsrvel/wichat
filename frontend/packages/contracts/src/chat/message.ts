import { z } from "zod";
import { mentionUserIdSchema, publicEntityIdSchema } from "../common";

export const MessageReplyPreviewSchema = z.object({
	id: publicEntityIdSchema,
	author_id: z.string(),
	author_display_name: z.string(),
	body_preview: z.string(),
});
export type MessageReplyPreview = z.infer<typeof MessageReplyPreviewSchema>;

export const MessageReactionSummarySchema = z.object({
	emoji: z.string(),
	count: z.number().int().nonnegative(),
	user_ids: z.array(z.string()).default([]),
});
export type MessageReactionSummary = z.infer<
	typeof MessageReactionSummarySchema
>;

export const MessageSchema = z.object({
	id: publicEntityIdSchema,
	channel_seq: z.number().int().positive().optional(),
	client_msg_id: z.string().uuid().optional(),
	author_id: z.string(),
	author_display_name: z.string(),
	author_avatar_url: z.string().default(""),
	body: z.string(),
	created_at: z.string(),
	edited_at: z.string().optional().default(""),
	is_pinned: z.boolean().default(false),
	reply_to: MessageReplyPreviewSchema.nullable().optional(),
	reactions: z.array(MessageReactionSummarySchema).default([]),
	mine: z.boolean().default(false),
	mentioned_user_ids: z.array(mentionUserIdSchema).optional().default([]),
});
export type Message = z.infer<typeof MessageSchema>;

export const MessagesListSchema = z.object({
	messages: z.array(MessageSchema),
	has_more: z.boolean().optional(),
});
export type MessagesList = z.infer<typeof MessagesListSchema>;

export const MessageChannelSeqSchema = z.object({
	message_id: publicEntityIdSchema,
	group_id: publicEntityIdSchema,
	channel_id: publicEntityIdSchema,
	channel_seq: z.number().int().positive(),
});
export type MessageChannelSeq = z.infer<typeof MessageChannelSeqSchema>;

export const CreateMessageBodySchema = z.object({
	body: z.string().trim().min(1).max(4096),
	reply_to_id: publicEntityIdSchema.optional(),
	client_msg_id: z.string().uuid().optional(),
	mentioned_user_ids: z.array(mentionUserIdSchema).optional().default([]),
});
export type CreateMessageBody = z.infer<typeof CreateMessageBodySchema>;

export const UpdateMessageBodySchema = z.object({
	body: z.string().trim().min(1).max(4096),
});
export type UpdateMessageBody = z.infer<typeof UpdateMessageBodySchema>;

export const ToggleReactionBodySchema = z.object({
	emoji: z.string().trim().min(1).max(64),
});
export type ToggleReactionBody = z.infer<typeof ToggleReactionBodySchema>;

export const CreateMessageResponseSchema = z.object({
	message: MessageSchema,
});
export type CreateMessageResponse = z.infer<typeof CreateMessageResponseSchema>;

export const ReactionsResponseSchema = z.object({
	reactions: z.array(MessageReactionSummarySchema).default([]),
});
export type ReactionsResponse = z.infer<typeof ReactionsResponseSchema>;

export const MessageAttachmentSchema = z.object({
	id: z.string().min(1),
	url: z.string().min(1),
	filename: z.string().min(1),
	content_type: z.string().min(1),
	size_bytes: z.number().int().nonnegative(),
});
export type MessageAttachment = z.infer<typeof MessageAttachmentSchema>;

export const UploadAttachmentResponseSchema = z.object({
	attachment: MessageAttachmentSchema,
});
export type UploadAttachmentResponse = z.infer<
	typeof UploadAttachmentResponseSchema
>;

export const RealtimeMessageCreatedSchema = z.object({
	type: z.literal("message.created"),
	message: MessageSchema,
});

export const RealtimeMessageUpdatedSchema = z.object({
	type: z.literal("message.updated"),
	message: MessageSchema,
});

export const RealtimeMessageDeletedSchema = z.object({
	type: z.literal("message.deleted"),
	message_id: publicEntityIdSchema,
});

export const RealtimeMessageReactionsUpdatedSchema = z.object({
	type: z.literal("message.reactions.updated"),
	message_id: publicEntityIdSchema,
	reactions: z.array(MessageReactionSummarySchema).default([]),
});

export const RealtimeChatEventSchema = z.discriminatedUnion("type", [
	RealtimeMessageCreatedSchema,
	RealtimeMessageUpdatedSchema,
	RealtimeMessageDeletedSchema,
	RealtimeMessageReactionsUpdatedSchema,
]);
export type RealtimeChatEvent = z.infer<typeof RealtimeChatEventSchema>;
