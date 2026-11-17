import { z } from "zod";
import { publicEntityIdSchema } from "../common";
import { ChannelCategorySchema } from "./channel-category";

export const channelSidebarSectionSchema = z.enum(["favorites", "channels"]);

export const channelKindSchema = z.enum([
	"broadcast",
	"text_only",
	"voice_only",
	"text_and_voice",
	"forum",
	"dm",
	"public",
	"private",
]);

export const channelContentVisibilitySchema = z.enum([
	"default",
	"spoiler",
	"age_restricted",
]);
export type ChannelContentVisibility = z.infer<
	typeof channelContentVisibilitySchema
>;

export const forumDefaultLayoutSchema = z.enum(["list", "gallery"]);
export type ForumDefaultLayout = z.infer<typeof forumDefaultLayoutSchema>;

export const forumSortOrderSchema = z.enum([
	"recent_activity",
	"creation_time",
]);
export type ForumSortOrder = z.infer<typeof forumSortOrderSchema>;

export const forumTagMatchingSchema = z.enum(["match_some", "match_all"]);
export type ForumTagMatching = z.infer<typeof forumTagMatchingSchema>;

export const voiceVideoQualitySchema = z.enum(["auto", "720p"]);
export type VoiceVideoQuality = z.infer<typeof voiceVideoQualitySchema>;

export const ChannelSchema = z.object({
	id: publicEntityIdSchema,
	name: z.string(),
	type: z.union([channelKindSchema, z.string()]),
	sidebar_section: channelSidebarSectionSchema.default("channels"),
	image_url: z.string().default(""),
	is_private: z.boolean().default(false),
	category_id: publicEntityIdSchema.or(z.literal("")).nullable().optional(),
	position: z.number().int().default(0),
	slowmode_seconds: z.number().int().default(0),
	thread_slowmode_seconds: z.number().int().default(0),
	content_visibility: channelContentVisibilitySchema.default("default"),
	forum_default_layout: forumDefaultLayoutSchema.default("list"),
	forum_sort_order: forumSortOrderSchema.default("recent_activity"),
	forum_tag_matching: forumTagMatchingSchema.default("match_some"),
	bitrate_kbps: z.number().int().default(64),
	video_quality: voiceVideoQualitySchema.default("auto"),
	user_limit: z.number().int().default(0),
	rtc_region: z.string().default("automatic"),
	caller_permissions: z.array(z.string()).default([]),
});
export type Channel = z.infer<typeof ChannelSchema>;

export const UpdateGroupChannelRequestSchema = z
	.object({
		name: z.string().trim().min(1).max(64).optional(),
		is_private: z.boolean().optional(),
		slowmode_seconds: z.number().int().min(0).max(21600).optional(),
		thread_slowmode_seconds: z.number().int().min(0).max(21600).optional(),
		content_visibility: channelContentVisibilitySchema.optional(),
		forum_default_layout: forumDefaultLayoutSchema.optional(),
		forum_sort_order: forumSortOrderSchema.optional(),
		forum_tag_matching: forumTagMatchingSchema.optional(),
		bitrate_kbps: z.number().int().min(8).max(96).optional(),
		video_quality: voiceVideoQualitySchema.optional(),
		user_limit: z.number().int().min(0).max(99).optional(),
		rtc_region: z.string().trim().min(1).max(64).optional(),
	})
	.refine((v) => Object.values(v).some((val) => val !== undefined));
export type UpdateGroupChannelRequest = z.infer<
	typeof UpdateGroupChannelRequestSchema
>;

export const UpdateGroupChannelResponseSchema = z.object({
	channel: ChannelSchema,
});
export type UpdateGroupChannelResponse = z.infer<
	typeof UpdateGroupChannelResponseSchema
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
	"text_and_voice",
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
