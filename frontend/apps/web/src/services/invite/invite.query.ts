import { queryOptions } from "@tanstack/react-query";
import { getGroupInviteLink, listGroupMembers } from "@/services/invite/invite.service";

export const Q_INVITE_MEMBERS = "invite.members";
export const Q_INVITE_LINK = "invite.link";
const MEMBER_PAGE_SIZE = 10;

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

export function groupInviteLinkQueryOpts(groupId: string, channelId?: string) {
	return queryOptions({
		queryKey: [Q_INVITE_LINK, groupId, channelId ?? ""],
		queryFn: () => getGroupInviteLink(groupId, channelId),
		staleTime: 60_000,
		retry: 1,
		enabled: Boolean(groupId),
	});
}
