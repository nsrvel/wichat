import { z } from "zod";
import { GroupMemberSchema } from "./group-member";

export const GroupRoleSchema = z.object({
	id: z.string(),
	slug: z.string(),
	name: z.string(),
	color: z.string().default("#99aab5"),
	hoist: z.boolean().default(false),
	mentionable: z.boolean().default(false),
	is_default: z.boolean().default(false),
	is_system: z.boolean().default(false),
	position: z.number().int().nonnegative().default(0),
	permissions: z.array(z.string()).default([]),
	member_count: z.number().int().nonnegative().default(0),
	created_at: z.string().optional().default(""),
	updated_at: z.string().optional().default(""),
});
export type GroupRole = z.infer<typeof GroupRoleSchema>;

export const GroupRolesListResponseSchema = z.object({
	roles: z.array(GroupRoleSchema),
	hide_roles: z.boolean().default(false),
});
export type GroupRolesListResponse = z.infer<
	typeof GroupRolesListResponseSchema
>;

export const GroupRoleResponseSchema = z.object({
	role: GroupRoleSchema,
});
export type GroupRoleResponse = z.infer<typeof GroupRoleResponseSchema>;

export const CreateGroupRoleRequestSchema = z.object({
	name: z.string().trim().max(48).optional().default(""),
});
export type CreateGroupRoleRequest = z.infer<
	typeof CreateGroupRoleRequestSchema
>;

export const UpdateGroupRoleRequestSchema = z.object({
	name: z.string().trim().min(1).max(48).optional(),
	color: z.string().trim().optional(),
	hoist: z.boolean().optional(),
	mentionable: z.boolean().optional(),
	permissions: z.array(z.string()).optional(),
});
export type UpdateGroupRoleRequest = z.infer<
	typeof UpdateGroupRoleRequestSchema
>;

export const GroupRoleMembersResponseSchema = z.object({
	members: z.array(GroupMemberSchema),
});
export type GroupRoleMembersResponse = z.infer<
	typeof GroupRoleMembersResponseSchema
>;

export const AddGroupRoleMembersRequestSchema = z.object({
	user_ids: z.array(z.string().min(1)).min(1),
});
export type AddGroupRoleMembersRequest = z.infer<
	typeof AddGroupRoleMembersRequestSchema
>;
