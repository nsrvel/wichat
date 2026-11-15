import { z } from "zod";

export const PresenceStatusSchema = z.enum([
	"online",
	"idle",
	"dnd",
	"offline",
]);
export type PresenceStatus = z.infer<typeof PresenceStatusSchema>;

export const ManualPresenceStatusSchema = z.enum([
	"online",
	"idle",
	"dnd",
	"invisible",
]);
export type ManualPresenceStatus = z.infer<typeof ManualPresenceStatusSchema>;

export const MyPresenceResponseSchema = z.object({
	manual_status: ManualPresenceStatusSchema,
	effective_status: PresenceStatusSchema,
	is_online: z.boolean(),
	expires_at: z.string().optional().default(""),
});
export type MyPresenceResponse = z.infer<typeof MyPresenceResponseSchema>;

export const UpdateMyPresenceRequestSchema = z.object({
	status: ManualPresenceStatusSchema,
	ttl_seconds: z.number().int().min(0).optional(),
});
export type UpdateMyPresenceRequest = z.infer<
	typeof UpdateMyPresenceRequestSchema
>;

export const RealtimePresenceUpdatedSchema = z.object({
	type: z.literal("presence.updated"),
	user_id: z.string(),
	presence_status: PresenceStatusSchema,
	is_online: z.boolean(),
});
export type RealtimePresenceUpdated = z.infer<
	typeof RealtimePresenceUpdatedSchema
>;
