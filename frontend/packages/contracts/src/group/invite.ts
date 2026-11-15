import { z } from "zod";
import { publicEntityIdSchema } from "../common";

export const GroupInviteLinkSchema = z.object({
	token: z.string(),
	url: z.string(),
	expires_at: z.string(),
});

export const SendGroupInvitesRequestSchema = z.object({
	user_ids: z.array(z.string()).default([]),
	emails: z.array(z.email()).default([]),
	channel_id: z.string().optional(),
});

export const SendGroupInvitesResultSchema = z.object({
	added: z.number().int().nonnegative(),
	emailed: z.number().int().nonnegative(),
	skipped: z.number().int().nonnegative(),
});

export const InvitePreviewSchema = z
	.object({
		valid: z.boolean(),
		is_private: z.boolean().optional(),
		group_name: z.string().optional(),
		group_description: z.string().optional(),
		channel_name: z.string().optional(),
		group_logo_url: z.string().optional(),
		groupLogoUrl: z.string().optional(),
		group_banner_url: z.string().optional(),
		group_banner_color: z.string().optional(),
		member_count: z.number().int().nonnegative().optional(),
		online_count: z.number().int().nonnegative().optional(),
		expires_at: z.string().optional(),
		reason: z.string().optional(),
	})
	.transform(({ groupLogoUrl, group_logo_url, ...rest }) => ({
		...rest,
		group_logo_url: group_logo_url ?? groupLogoUrl,
	}));

export type InvitePreview = z.infer<typeof InvitePreviewSchema>;

export const RedeemInviteRequestSchema = z.object({
	token: z.string().trim().min(1),
});

export const RedeemInviteResultSchema = z.object({
	group_id: publicEntityIdSchema,
	channel_id: z.union([publicEntityIdSchema, z.literal("")]).optional(),
});
