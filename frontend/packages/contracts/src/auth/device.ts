import { z } from "zod";

export const DeviceSessionSchema = z.object({
	id: z.string(),
	platform: z.string().optional().default("web"),
	device_label: z.string().optional().default(""),
	location_label: z.string().optional().default(""),
	user_agent: z.string().optional().default(""),
	last_used_at: z.string(),
	created_at: z.string(),
	is_current: z.boolean(),
});

export type DeviceSession = z.infer<typeof DeviceSessionSchema>;

export const DeviceSessionsResponseSchema = z.object({
	devices: z.array(DeviceSessionSchema),
});

export type DeviceSessionsResponse = z.infer<
	typeof DeviceSessionsResponseSchema
>;

export const RevokeDeviceRequestSchema = z.object({
	current_password: z.string().optional().default(""),
});

export type RevokeDeviceRequest = z.infer<typeof RevokeDeviceRequestSchema>;
