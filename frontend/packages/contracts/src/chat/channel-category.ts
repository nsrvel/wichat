import { z } from "zod";
import { publicEntityIdSchema } from "../common";

export const ChannelCategorySchema = z.object({
	id: publicEntityIdSchema,
	name: z.string(),
	is_private: z.boolean().default(false),
	position: z.number().int().default(0),
});
export type ChannelCategory = z.infer<typeof ChannelCategorySchema>;

export const CreateChannelCategoryRequestSchema = z.object({
	name: z.string().trim().min(1).max(64),
	is_private: z.boolean().default(false),
});
export type CreateChannelCategoryRequest = z.infer<
	typeof CreateChannelCategoryRequestSchema
>;

export const CreateChannelCategoryResponseSchema = z.object({
	category: ChannelCategorySchema,
});
export type CreateChannelCategoryResponse = z.infer<
	typeof CreateChannelCategoryResponseSchema
>;

export const UpdateChannelCategoryRequestSchema = z
	.object({
		name: z.string().trim().min(1).max(64).optional(),
		is_private: z.boolean().optional(),
	})
	.refine((v) => v.name !== undefined || v.is_private !== undefined);
export type UpdateChannelCategoryRequest = z.infer<
	typeof UpdateChannelCategoryRequestSchema
>;

export const UpdateChannelCategoryResponseSchema = z.object({
	category: ChannelCategorySchema,
});
export type UpdateChannelCategoryResponse = z.infer<
	typeof UpdateChannelCategoryResponseSchema
>;

export const categoryPermissionTargetTypeSchema = z.enum(["role", "member"]);
export type CategoryPermissionTargetType = z.infer<
	typeof categoryPermissionTargetTypeSchema
>;

export const ChannelCategoryPermissionOverrideSchema = z.object({
	target_type: categoryPermissionTargetTypeSchema,
	target_id: publicEntityIdSchema,
	target_name: z.string(),
	target_color: z.string().default(""),
	target_avatar_url: z.string().default(""),
	is_everyone: z.boolean().default(false),
	allow_permissions: z.array(z.string()).default([]),
	deny_permissions: z.array(z.string()).default([]),
});
export type ChannelCategoryPermissionOverride = z.infer<
	typeof ChannelCategoryPermissionOverrideSchema
>;

export const CategoryPermissionsListSchema = z.object({
	overrides: z.array(ChannelCategoryPermissionOverrideSchema),
});
export type CategoryPermissionsList = z.infer<
	typeof CategoryPermissionsListSchema
>;

export const UpsertCategoryPermissionOverrideRequestSchema = z.object({
	target_type: categoryPermissionTargetTypeSchema,
	target_id: publicEntityIdSchema,
	allow_permissions: z.array(z.string()),
	deny_permissions: z.array(z.string()),
});
export type UpsertCategoryPermissionOverrideRequest = z.infer<
	typeof UpsertCategoryPermissionOverrideRequestSchema
>;

export const UpsertCategoryPermissionOverrideResponseSchema = z.object({
	override: ChannelCategoryPermissionOverrideSchema,
});
export type UpsertCategoryPermissionOverrideResponse = z.infer<
	typeof UpsertCategoryPermissionOverrideResponseSchema
>;

export const ReorderChannelCategoriesRequestSchema = z.object({
	category_ids: z.array(publicEntityIdSchema).min(1),
});
export type ReorderChannelCategoriesRequest = z.infer<
	typeof ReorderChannelCategoriesRequestSchema
>;
