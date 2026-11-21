import type { GroupMember } from "@wichat/contracts/group/group-member";
import type { GroupMembersPage } from "@wichat/contracts/group/group-member";
import type { MyPresenceResponse, PresenceStatus } from "@wichat/contracts/user/presence";
import type { QueryClient } from "@tanstack/react-query";
import { Q_GROUP_ROLE_MEMBERS } from "@/services/group-role/group-role.key";
import { Q_INVITE_MEMBERS, Q_INVITE_PREVIEW } from "@/services/invite/invite.query";
import { Q_MY_PRESENCE } from "@/services/user/user.key";

export type LivePresenceFrame = {
	user_id: string;
	presence_status: PresenceStatus;
	is_online: boolean;
};

function patchMemberRow(member: GroupMember, frame: LivePresenceFrame): GroupMember | null {
	if (member.user_id !== frame.user_id) {
		return null;
	}
	if (member.is_online === frame.is_online && member.presence_status === frame.presence_status) {
		return member;
	}
	return {
		...member,
		is_online: frame.is_online,
		presence_status: frame.presence_status,
	};
}

function patchPage(
	prev: GroupMembersPage | undefined,
	userId: string,
	frame: LivePresenceFrame,
): GroupMembersPage | undefined {
	if (!prev) {
		return prev;
	}
	let changed = false;
	const members = prev.members.map((m) => {
		const next = patchMemberRow(m, frame);
		if (!next || next === m) {
			return m;
		}
		changed = true;
		return next;
	});
	return changed ? { ...prev, members } : prev;
}

function patchMemberArray(
	prev: GroupMember[] | undefined,
	frame: LivePresenceFrame,
): GroupMember[] | undefined {
	if (!prev?.length) {
		return prev;
	}
	let changed = false;
	const members = prev.map((m) => {
		const next = patchMemberRow(m, frame);
		if (!next || next === m) {
			return m;
		}
		changed = true;
		return next;
	});
	return changed ? members : prev;
}

/** Unique online members across all cached member-list pages for a group. */
export function onlineCountFromMemberCache(queryClient: QueryClient, groupId: string): number {
	const online = new Map<string, boolean>();
	for (const [, page] of queryClient.getQueriesData<GroupMembersPage>({
		queryKey: [Q_INVITE_MEMBERS, groupId],
	})) {
		for (const m of page?.members ?? []) {
			online.set(m.user_id, Boolean(m.is_online));
		}
	}
	let count = 0;
	for (const isOn of online.values()) {
		if (isOn) {
			count += 1;
		}
	}
	return count;
}

let invitePreviewRefreshTimer: ReturnType<typeof setTimeout> | undefined;

/** Debounced HTTP refresh for invite landing `online_count` (preview has no group id for WS math). */
export function scheduleInvitePreviewPresenceRefresh(queryClient: QueryClient): void {
	if (invitePreviewRefreshTimer) {
		clearTimeout(invitePreviewRefreshTimer);
	}
	invitePreviewRefreshTimer = setTimeout(() => {
		invitePreviewRefreshTimer = undefined;
		void queryClient.invalidateQueries({ queryKey: [Q_INVITE_PREVIEW] });
	}, 2_000);
}

export function patchLiveMemberPresence(
	queryClient: QueryClient,
	frame: LivePresenceFrame,
	opts?: { refreshInvitePreview?: boolean },
): void {
	queryClient.setQueriesData<GroupMembersPage>({ queryKey: [Q_INVITE_MEMBERS] }, (prev) =>
		patchPage(prev, frame.user_id, frame),
	);
	queryClient.setQueriesData<GroupMember[]>({ queryKey: [Q_GROUP_ROLE_MEMBERS] }, (prev) =>
		patchMemberArray(prev, frame),
	);
	if (opts?.refreshInvitePreview) {
		scheduleInvitePreviewPresenceRefresh(queryClient);
	}
}

export function upsertLiveMyPresence(
	queryClient: QueryClient,
	frame: LivePresenceFrame,
	selfUserId: string,
): void {
	if (frame.user_id !== selfUserId) {
		return;
	}
	queryClient.setQueryData<MyPresenceResponse>([Q_MY_PRESENCE], (prev) => ({
		manual_status: prev?.manual_status ?? "online",
		effective_status: frame.presence_status,
		is_online: frame.is_online,
		expires_at: prev?.expires_at ?? "",
	}));
}