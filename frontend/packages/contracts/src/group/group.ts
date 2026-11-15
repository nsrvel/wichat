import { z } from "zod";
import { groupNameSchema, publicEntityIdSchema } from "../common";

export const groupMembershipRoleSchema = z.enum(["owner", "member"]);

export const GroupSchema = z.object({
	id: publicEntityIdSchema,
	name: z.string(),
	logo_url: z.string(),
	description: z.string().default(""),
	is_private: z.boolean().default(false),
	banner_color: z.string().default(""),
	banner_url: z.string().default(""),
	hide_roles: z.boolean().default(false),
	default_channel_id: z.union([publicEntityIdSchema, z.literal("")]),
	role: z.preprocess(
		(v) => (v === "" || v == null ? undefined : v),
		groupMembershipRoleSchema.or(z.enum(["admin", "guest"])).optional(),
	),
	effective_permissions: z.array(z.string()).default([]),
});
export type Group = z.infer<typeof GroupSchema>;

export const GroupsListSchema = z.object({
	groups: z.array(GroupSchema),
});
export type GroupsList = z.infer<typeof GroupsListSchema>;

/** Groups the user belongs to — membership role is always set. */
export const MyGroupSchema = GroupSchema.extend({
	role: groupMembershipRoleSchema,
});

export const MyGroupsListSchema = z.object({
	groups: z.array(MyGroupSchema),
});

export const ReorderMyGroupsRequestSchema = z.object({
	group_ids: z.array(publicEntityIdSchema).min(1),
});

export const CreateGroupRequestSchema = z.object({
	name: groupNameSchema,
});
export type CreateGroupRequest = z.infer<typeof CreateGroupRequestSchema>;

export const UpdateGroupProfileBodySchema = z.object({
	name: groupNameSchema,
	description: z.string().trim().max(2000),
	is_private: z.boolean(),
	banner_color: z.string(),
	clear_logo: z.boolean(),
	clear_banner_image: z.boolean(),
});
export type UpdateGroupProfileBody = z.infer<
	typeof UpdateGroupProfileBodySchema
>;

export const UpdateGroupHideRolesRequestSchema = z.object({
	hide_roles: z.boolean(),
});
export type UpdateGroupHideRolesRequest = z.infer<
	typeof UpdateGroupHideRolesRequestSchema
>;

export const UpdateGroupHideRolesResponseSchema = z.object({
	hide_roles: z.boolean(),
});
export type UpdateGroupHideRolesResponse = z.infer<
	typeof UpdateGroupHideRolesResponseSchema
>;
