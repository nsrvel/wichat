import { z } from "zod";
import { PresenceStatusSchema } from "../user/presence";
import { groupMembershipRoleSchema } from "./group";

export const GroupMemberRoleSummarySchema = z.object({
	id: z.string(),
	slug: z.string(),
	name: z.string(),
	color: z.string(),
	hoist: z.boolean().optional().default(false),
	position: z.number().int().optional().default(0),
});
export type GroupMemberRoleSummary = z.infer<
	typeof GroupMemberRoleSummarySchema
>;

export const MemberJoinMethodSchema = z.enum([
	"owner",
	"open",
	"invite",
	"invite_link",
	"unknown",
]);
export type MemberJoinMethod = z.infer<typeof MemberJoinMethodSchema>;

export const GroupMemberSchema = z.object({
	user_id: z.string(),
	display_name: z.string(),
	username: z.string(),
	avatar_url: z.string().optional().default(""),
	role: groupMembershipRoleSchema,
	joined_at: z.string(),
	roles: z.array(GroupMemberRoleSummarySchema).optional().default([]),
	is_online: z.boolean().optional().default(false),
	presence_status: PresenceStatusSchema.optional().default("offline"),
	banner_color: z.string().optional().default(""),
	bio: z.string().optional().default(""),
	pronouns: z.string().optional().default(""),
	user_created_at: z.string().optional().default(""),
	status_note_text: z.string().optional().default(""),
	status_note_emoji: z.string().optional().default(""),
	status_note_expires_at: z.string().nullable().optional().default(null),
	join_method: MemberJoinMethodSchema.optional().default("unknown"),
	invite_code: z.string().optional().default(""),
	invited_by_user_id: z.string().optional().default(""),
	invited_by_username: z.string().optional().default(""),
	invited_by_display_name: z.string().optional().default(""),
	timeout_until: z.string().nullable().optional().default(null),
});
export type GroupMember = z.infer<typeof GroupMemberSchema>;

export const GroupMembersPageSchema = z.object({
	members: z.array(GroupMemberSchema),
	next_cursor: z.string().optional().default(""),
	total: z.number().int().nonnegative(),
	/** Gateway signals ms-user outage — clients should keep cached member rows. */
	degraded: z.boolean().optional().default(false),
});
export type GroupMembersPage = z.infer<typeof GroupMembersPageSchema>;

export const KickGroupMemberRequestSchema = z.object({
	reason: z.string().max(512).optional().default(""),
});
export type KickGroupMemberRequest = z.infer<
	typeof KickGroupMemberRequestSchema
>;

export const BanGroupMembersRequestSchema = z.object({
	user_ids: z.array(z.string().min(1)).min(1),
	reason: z.string().max(512).optional().default(""),
	delete_message_seconds: z
		.number()
		.int()
		.min(0)
		.max(604800)
		.optional()
		.default(0),
	duration_seconds: z.number().int().min(0).optional().default(0),
});
export type BanGroupMembersRequest = z.infer<
	typeof BanGroupMembersRequestSchema
>;

export const BanGroupMembersResponseSchema = z.object({
	ok: z.boolean(),
	banned_count: z.number().int().nonnegative(),
});
export type BanGroupMembersResponse = z.infer<
	typeof BanGroupMembersResponseSchema
>;

export const TimeoutGroupMemberRequestSchema = z.object({
	duration_seconds: z.number().int().min(0).max(604800),
	reason: z.string().max(512).optional().default(""),
});
export type TimeoutGroupMemberRequest = z.infer<
	typeof TimeoutGroupMemberRequestSchema
>;

export const TimeoutGroupMemberResponseSchema = z.object({
	timeout_until: z.string().optional().default(""),
});
export type TimeoutGroupMemberResponse = z.infer<
	typeof TimeoutGroupMemberResponseSchema
>;

export const GroupBanEntrySchema = z.object({
	user_id: z.string().min(1),
	username: z.string(),
	display_name: z.string(),
	avatar_url: z.string().optional().default(""),
	reason: z.string().optional().default(""),
	banned_by_user_id: z.string().optional().default(""),
	banned_by_username: z.string().optional().default(""),
	banned_by_display_name: z.string().optional().default(""),
	expires_at: z.string().optional().default(""),
	created_at: z.string(),
});
export type GroupBanEntry = z.infer<typeof GroupBanEntrySchema>;

export const GroupBansPageSchema = z.object({
	bans: z.array(GroupBanEntrySchema),
	next_cursor: z.string().optional().default(""),
	total: z.number().int().nonnegative(),
});
export type GroupBansPage = z.infer<typeof GroupBansPageSchema>;
