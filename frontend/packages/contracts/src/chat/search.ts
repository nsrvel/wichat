import { z } from "zod";

export const MessageSearchHitSchema = z.object({
	message_id: z.string(),
	group_id: z.string(),
	channel_id: z.string(),
	body_preview: z.string(),
	author_display_name: z.string(),
	created_at: z.string(),
});

export const MessageSearchHitsSchema = z.object({
	hits: z.array(MessageSearchHitSchema),
});

export type MessageSearchHit = z.infer<typeof MessageSearchHitSchema>;
export type MessageSearchHits = z.infer<typeof MessageSearchHitsSchema>;

export const FileSearchHitSchema = z.object({
	attachment_id: z.string(),
	message_id: z.string(),
	group_id: z.string(),
	channel_id: z.string(),
	filename: z.string(),
	content_type: z.string(),
	created_at: z.string(),
});

export const FileSearchHitsSchema = z.object({
	hits: z.array(FileSearchHitSchema),
});

export type FileSearchHit = z.infer<typeof FileSearchHitSchema>;
export type FileSearchHits = z.infer<typeof FileSearchHitsSchema>;
