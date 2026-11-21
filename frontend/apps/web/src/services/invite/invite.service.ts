import {
	GroupInviteLinkSchema,
	InvitePreviewSchema,
	RedeemInviteRequestSchema,
	RedeemInviteResultSchema,
	SendGroupInvitesRequestSchema,
	SendGroupInvitesResultSchema,
} from "@wichat/contracts/group/invite";
import { GroupMembersPageSchema } from "@wichat/contracts/group/member";
import { UserSearchSchema } from "@wichat/contracts/user";
import { ApiError } from "@wichat/api-client/api-error";
import { api } from "@/integrations/api-client";

export async function listGroupMembers(
	groupId: string,
	opts?: { limit?: number; cursor?: string },
) {
	const params = new URLSearchParams();
	if (opts?.limit) {
		params.set("limit", String(opts.limit));
	}
	if (opts?.cursor) {
		params.set("cursor", opts.cursor);
	}
	const qs = params.toString();
	const path = `/groups/${encodeURIComponent(groupId)}/members${qs ? `?${qs}` : ""}`;
	const page = await api.get(path, GroupMembersPageSchema);
	if (page.degraded) {
		throw new ApiError("error.service_unavailable");
	}
	return page;
}

export async function searchUsers(
	query: string,
	opts?: { excludeGroupId?: string; limit?: number },
) {
	const params = new URLSearchParams({ q: query });
	if (opts?.excludeGroupId) {
		params.set("exclude_group_id", opts.excludeGroupId);
	}
	if (opts?.limit) {
		params.set("limit", String(opts.limit));
	}
	return api.get(`/users/search?${params.toString()}`, UserSearchSchema);
}

export async function getGroupInviteLink(
	groupId: string,
	channelId?: string,
	turnstileToken?: string,
) {
	const params = new URLSearchParams();
	if (channelId) {
		params.set("channel_id", channelId);
	}
	const captcha = turnstileToken?.trim();
	if (captcha) {
		params.set("turnstile_token", captcha);
	}
	const qs = params.toString();
	const path = `/groups/${encodeURIComponent(groupId)}/invite-link${qs ? `?${qs}` : ""}`;
	return api.get(path, GroupInviteLinkSchema);
}

export async function sendGroupInvites(
	groupId: string,
	body: {
		user_ids: string[];
		emails: string[];
		channel_id?: string;
		turnstile_token?: string;
	},
) {
	const payload = SendGroupInvitesRequestSchema.parse(body);
	return api.post(
		`/groups/${encodeURIComponent(groupId)}/invites`,
		SendGroupInvitesResultSchema,
		payload,
	);
}

export async function previewInvite(token: string, turnstileToken?: string) {
	const params = new URLSearchParams({ invite: token });
	const captcha = turnstileToken?.trim();
	if (captcha) {
		params.set("turnstile_token", captcha);
	}
	return api.get(`/invites/preview?${params.toString()}`, InvitePreviewSchema);
}

export async function redeemInvite(token: string, turnstileToken?: string) {
	const body = RedeemInviteRequestSchema.parse({
		token,
		turnstile_token: turnstileToken?.trim() || undefined,
	});
	return api.post("/invites/redeem", RedeemInviteResultSchema, body);
}
