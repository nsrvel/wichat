import { z } from "zod";

export const UserSchema = z.object({
	id: z.string(),
	email: z.string(),
	display_name: z.string(),
});

export const SessionSchema = z.object({
	access_token: z.string(),
	expires_in: z.number(),
	user: UserSchema,
});

export const SetupStatusSchema = z.object({
	needs_setup: z.boolean(),
});

export const OkSchema = z.object({
	ok: z.boolean(),
});

export const LoginRequestSchema = z.object({
	email: z.email(),
	password: z.string().min(12).max(128),
});

export const SetupRequestSchema = LoginRequestSchema.extend({
	display_name: z.string().trim().min(1).max(80),
});

export type User = z.infer<typeof UserSchema>;
export type Session = z.infer<typeof SessionSchema>;
export type SetupStatus = z.infer<typeof SetupStatusSchema>;
export type LoginRequest = z.infer<typeof LoginRequestSchema>;
export type SetupRequest = z.infer<typeof SetupRequestSchema>;
