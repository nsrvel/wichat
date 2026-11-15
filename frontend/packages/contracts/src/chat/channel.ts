import { z } from "zod";
import { publicEntityIdSchema } from "../common";

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

export const ChannelSchema = z.object({
	id: publicEntityIdSchema,
	name: z.string(),
	type: z.union([channelKindSchema, z.string()]),
	sidebar_section: channelSidebarSectionSchema.default("channels"),
	image_url: z.string().default(""),
	is_private: z.boolean().default(false),
});
export type Channel = z.infer<typeof ChannelSchema>;

export const ChannelsListSchema = z.object({
	channels: z.array(ChannelSchema),
});
export type ChannelsList = z.infer<typeof ChannelsListSchema>;
