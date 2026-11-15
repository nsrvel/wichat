import { z } from "zod";

export const ActivityFieldChangeSchema = z.object({
	old: z.unknown().optional(),
	new: z.unknown().optional(),
});

export type ActivityFieldChange = z.infer<typeof ActivityFieldChangeSchema>;

export const UserActivityLogSchema = z.object({
	id: z.string(),
	user_id: z.string(),
	actor_id: z.string(),
	category: z.string(),
	action: z.string(),
	target_type: z.string(),
	target_id: z.string(),
	summary: z.string(),
	changes: z.record(z.string(), ActivityFieldChangeSchema).default({}),
	metadata: z.record(z.string(), z.unknown()).default({}),
	created_at: z.string(),
});

export type UserActivityLog = z.infer<typeof UserActivityLogSchema>;

export const ListUserActivityLogsResponseSchema = z.object({
	logs: z.array(UserActivityLogSchema),
	next_cursor: z.string().optional(),
});

export type ListUserActivityLogsResponse = z.infer<
	typeof ListUserActivityLogsResponseSchema
>;
