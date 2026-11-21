import { queryOptions } from "@tanstack/react-query";
import { writeInvitePreviewCache } from "@/services/invite/invite-preview-cache";
import {
	getGroupInviteLink,
	listGroupMembers,
	previewInvite,
} from "@/services/invite/invite.service";

export const Q_INVITE_MEMBERS = "invite.members";
export const Q_INVITE_LINK = "invite.link";
export const Q_INVITE_PREVIEW = "invite.preview";
export const MEMBER_PAGE_SIZE = 10;

export function invitePreviewQueryKey(inviteToken: string, turnstileToken?: string) {
	return [Q_INVITE_PREVIEW, inviteToken.trim(), turnstileToken ?? ""] as const;
}

export function groupMembersQueryOpts(groupId: string, cursor = "") {
	return queryOptions({
		queryKey: [Q_INVITE_MEMBERS, groupId, cursor],
		queryFn: () =>
			listGroupMembers(groupId, {
				limit: MEMBER_PAGE_SIZE,
				cursor: cursor || undefined,
			}),
		staleTime: 30_000,
		retry: 1,
		enabled: Boolean(groupId),
	});
}

export function groupInviteLinkQueryOpts(
	groupId: string,
	channelId?: string,
	turnstileToken?: string,
	captchaReady = true,
) {
	return queryOptions({
		queryKey: [Q_INVITE_LINK, groupId, channelId ?? "", turnstileToken ?? ""],
		queryFn: () => getGroupInviteLink(groupId, channelId, turnstileToken),
		staleTime: 60_000,
		retry: 1,
		enabled: Boolean(groupId) && captchaReady,
	});
}

export function invitePreviewQueryOpts(
	inviteToken: string,
	turnstileToken?: string,
	captchaReady = true,
) {
	const token = inviteToken.trim();
	return queryOptions({
		queryKey: invitePreviewQueryKey(token, turnstileToken),
		queryFn: async () => {
			const next = await previewInvite(token, turnstileToken);
			writeInvitePreviewCache(token, next);
			return next;
		},
		staleTime: 30_000,
		gcTime: 10 * 60_000,
		retry: 1,
		refetchOnWindowFocus: false,
		enabled: Boolean(token) && captchaReady,
	});
}
