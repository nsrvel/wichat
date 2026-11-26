import { z } from "zod";
import { publicEntityIdSchema } from "../common";

export const GroupMentionHitSchema = z.object({
	message_id: publicEntityIdSchema,
	channel_id: publicEntityIdSchema,
	channel_name: z.string(),
	channel_type: z.string(),
	author_display_name: z.string(),
	body_preview: z.string(),
	created_at: z.string(),
	is_unread: z.boolean().default(false),
});
export type GroupMentionHit = z.infer<typeof GroupMentionHitSchema>;

export const GroupMentionsListSchema = z.object({
	hits: z.array(GroupMentionHitSchema).default([]),
	next_cursor: z.string().optional().default(""),
});
export type GroupMentionsList = z.infer<typeof GroupMentionsListSchema>;

export const MentionsUnreadCountSchema = z.object({
	count: z.number().int().nonnegative(),
});
export type MentionsUnreadCount = z.infer<typeof MentionsUnreadCountSchema>;

export const MarkMentionsReadBodySchema = z.object({
	message_ids: z.array(publicEntityIdSchema).min(1),
});
export type MarkMentionsReadBody = z.infer<typeof MarkMentionsReadBodySchema>;
