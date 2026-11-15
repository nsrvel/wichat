import { z } from "zod";
import { usernameSchema } from "../common";

export const UserSchema = z.object({
	id: z.string(),
	email: z.string(),
	display_name: z.string(),
	username: z.string(),
	avatar_url: z.string().optional().default(""),
	pronouns: z.string().optional().default(""),
	bio: z.string().optional().default(""),
	banner_color: z.string().optional().default(""),
	status_note_text: z.string().optional().default(""),
	status_note_emoji: z.string().optional().default(""),
	status_note_expires_at: z.string().nullable().optional().default(null),
	locale: z.string().optional().default(""),
	time_format: z.enum(["auto", "12h", "24h"]).optional().default("auto"),
	has_password: z.boolean().default(true),
	created_at: z.string().optional().default(""),
});

export type User = z.infer<typeof UserSchema>;

export const UpdateUserProfileRequestSchema = z.object({
	display_name: z.string(),
	pronouns: z.string().optional().default(""),
	bio: z.string().optional().default(""),
	banner_color: z.string().optional().default(""),
	clear_avatar: z.boolean().optional().default(false),
});

export type UpdateUserProfileRequest = z.infer<
	typeof UpdateUserProfileRequestSchema
>;

export const UpdateUserStatusNoteRequestSchema = z.object({
	text: z.string().max(80).optional().default(""),
	emoji: z.string().max(16).optional().default(""),
	expires_at: z.string().nullable().optional().default(null),
});

export type UpdateUserStatusNoteRequest = z.infer<
	typeof UpdateUserStatusNoteRequestSchema
>;

export const UpdateUserPreferencesRequestSchema = z.object({
	locale: z.enum(["en", "id"]),
	time_format: z.enum(["auto", "12h", "24h"]),
});

export type UpdateUserPreferencesRequest = z.infer<
	typeof UpdateUserPreferencesRequestSchema
>;

export const UpdateUsernameRequestSchema = z.object({
	username: usernameSchema,
	current_password: z.string().min(1),
});

export type UpdateUsernameRequest = z.infer<typeof UpdateUsernameRequestSchema>;

export const RequestEmailChangeRequestSchema = z.object({
	locale: z.string().optional().default("en"),
});

export type RequestEmailChangeRequest = z.infer<
	typeof RequestEmailChangeRequestSchema
>;

export const VerifyEmailChangeRequestSchema = z.object({
	challenge_id: z.string().trim().min(1),
	code: z.string().trim().min(1).max(16),
});

export type VerifyEmailChangeRequest = z.infer<
	typeof VerifyEmailChangeRequestSchema
>;

export const VerifyEmailChangeResponseSchema = z.object({
	change_token: z.string().min(1),
});

export type VerifyEmailChangeResponse = z.infer<
	typeof VerifyEmailChangeResponseSchema
>;

export const CompleteEmailChangeRequestSchema = z.object({
	change_token: z.string().trim().min(1),
	new_email: z.email(),
	current_password: z.string().min(1),
	reason: z.string().trim().optional().default(""),
});

export type CompleteEmailChangeRequest = z.infer<
	typeof CompleteEmailChangeRequestSchema
>;

export const UpdatePasswordRequestSchema = z.object({
	current_password: z.string().optional().default(""),
	new_password: z.string().min(12).max(128),
});

export type UpdatePasswordRequest = z.infer<typeof UpdatePasswordRequestSchema>;

export const AccountActionRequestSchema = z.object({
	current_password: z.string().optional().default(""),
});

export type AccountActionRequest = z.infer<typeof AccountActionRequestSchema>;

export const UserSearchHitSchema = z.object({
	user_id: z.string(),
	display_name: z.string(),
	username: z.string(),
	avatar_url: z.string().optional().default(""),
});

export const UserSearchSchema = z.object({
	users: z.array(UserSearchHitSchema),
});
