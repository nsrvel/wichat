import { z } from "zod";
import { publicEntityIdSchema } from "../common";

export const MessageSchema = z.object({
	id: publicEntityIdSchema,
	author_id: z.string(),
	author_display_name: z.string(),
	body: z.string(),
	created_at: z.string(),
	mine: z.boolean().default(false),
});
export type Message = z.infer<typeof MessageSchema>;

export const MessagesListSchema = z.object({
	messages: z.array(MessageSchema),
});
export type MessagesList = z.infer<typeof MessagesListSchema>;

export const CreateMessageBodySchema = z.object({
	body: z.string().trim().min(1).max(4096),
});
export type CreateMessageBody = z.infer<typeof CreateMessageBodySchema>;

export const CreateMessageResponseSchema = z.object({
	message: MessageSchema,
});
export type CreateMessageResponse = z.infer<typeof CreateMessageResponseSchema>;

export const RealtimeMessageCreatedSchema = z.object({
	type: z.literal("message.created"),
	message: MessageSchema,
});
