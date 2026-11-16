import { z } from "zod";
import { publicEntityIdSchema } from "../common";

export const channelSidebarSectionSchema = z.enum(["favorites", "channels"]);

export const channelKindSchema = z.enum([
	"broadcast",
	"text_only",
	"voice_only",
	"forum",
	"dm",
	"public",
	"private",
]);

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

export const ChannelSchema = z.object({
	id: publicEntityIdSchema,
	name: z.string(),
	type: z.union([channelKindSchema, z.string()]),
	sidebar_section: channelSidebarSectionSchema.default("channels"),
	image_url: z.string().default(""),
	is_private: z.boolean().default(false),
	category_id: publicEntityIdSchema.or(z.literal("")).nullable().optional(),
	position: z.number().int().default(0),
});
export type Channel = z.infer<typeof ChannelSchema>;

export const ReorderChannelCategoriesRequestSchema = z.object({
	category_ids: z.array(publicEntityIdSchema).min(1),
});
export type ReorderChannelCategoriesRequest = z.infer<
	typeof ReorderChannelCategoriesRequestSchema
>;

export const ReorderGroupChannelsRequestSchema = z.object({
	channel_ids: z.array(publicEntityIdSchema).min(1),
	category_id: publicEntityIdSchema.optional(),
	sidebar_section: channelSidebarSectionSchema.optional(),
});
export type ReorderGroupChannelsRequest = z.infer<
	typeof ReorderGroupChannelsRequestSchema
>;

export const ChannelsListSchema = z.object({
	channels: z.array(ChannelSchema),
	categories: z.array(ChannelCategorySchema).default([]),
});
export type ChannelsList = z.infer<typeof ChannelsListSchema>;

export const creatableChannelKindSchema = z.enum([
	"text_only",
	"voice_only",
	"forum",
	"broadcast",
]);
export type CreatableChannelKind = z.infer<typeof creatableChannelKindSchema>;

export const CreateGroupChannelRequestSchema = z.object({
	name: z.string().trim().min(1).max(64),
	type: creatableChannelKindSchema,
	category_id: publicEntityIdSchema.optional(),
	is_private: z.boolean().default(false),
});
export type CreateGroupChannelRequest = z.infer<
	typeof CreateGroupChannelRequestSchema
>;

export const CreateGroupChannelResponseSchema = z.object({
	channel: ChannelSchema,
});
export type CreateGroupChannelResponse = z.infer<
	typeof CreateGroupChannelResponseSchema
>;
