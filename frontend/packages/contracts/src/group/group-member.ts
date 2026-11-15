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
});
export type GroupMember = z.infer<typeof GroupMemberSchema>;

export const GroupMembersPageSchema = z.object({
	members: z.array(GroupMemberSchema),
	next_cursor: z.string().optional().default(""),
	total: z.number().int().nonnegative(),
});
export type GroupMembersPage = z.infer<typeof GroupMembersPageSchema>;

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
