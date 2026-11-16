import { z } from "zod";
import { ActivityFieldChangeSchema } from "../user/activity-log";

export const GroupActivityLogSchema = z.object({
	id: z.string(),
	group_id: z.string(),
	actor_id: z.string(),
	actor_username: z.string().default(""),
	actor_display_name: z.string().default(""),
	actor_avatar_url: z.string().default(""),
	category: z.string(),
	action: z.string(),
	target_type: z.string(),
	target_id: z.string(),
	target_name: z.string().default(""),
	summary: z.string(),
	changes: z.record(z.string(), ActivityFieldChangeSchema).default({}),
	metadata: z.record(z.string(), z.unknown()).default({}),
	created_at: z.string(),
});

export type GroupActivityLog = z.infer<typeof GroupActivityLogSchema>;

export const GroupActivityLogsPageSchema = z.object({
	logs: z.array(GroupActivityLogSchema),
	next_cursor: z.string().default(""),
});

export type GroupActivityLogsPage = z.infer<typeof GroupActivityLogsPageSchema>;
