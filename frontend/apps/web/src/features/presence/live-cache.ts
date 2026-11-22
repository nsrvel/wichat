import type { GroupMember, GroupMembersPage } from "@wichat/contracts/group/member";
import type {
	ManualPresenceStatus,
	MyPresenceResponse,
	PresenceStatus,
} from "@wichat/contracts/user/presence";
import type { QueryClient } from "@tanstack/react-query";
import { Q_GROUP_ROLE_MEMBERS } from "@/services/group-role/group-role.key";
import { Q_INVITE_MEMBERS, Q_INVITE_PREVIEW } from "@/services/invite/invite.query";
import { Q_MY_PRESENCE } from "@/services/user/user.key";

export type LivePresenceFrame = {
	user_id: string;
	presence_status: PresenceStatus;
	is_online: boolean;
};

/** WS heartbeat interval (25s) × margin — HTTP must not clobber fresher WS self state. */
const WS_PRESENCE_HTTP_GUARD_MS = 90_000;

type WsSelfPresenceStamp = {
	at: number;
	frame: LivePresenceFrame;
};

let lastWsSelfPresence: WsSelfPresenceStamp | null = null;
let pendingManualStatus: ManualPresenceStatus | undefined;

/** Optimistic status picker — survives until PUT settles or errors. */
export function setPendingManualPresence(status: ManualPresenceStatus): void {
	pendingManualStatus = status;
}

export function clearPendingManualPresence(): void {
	pendingManualStatus = undefined;
}

/** Test-only reset of module coordination state. */
export function resetPresenceMergeStateForTests(): void {
	lastWsSelfPresence = null;
	pendingManualStatus = undefined;
}

export function touchWsSelfPresence(frame: LivePresenceFrame): void {
	lastWsSelfPresence = { at: Date.now(), frame };
}

function wsStampFresh(now = Date.now()): WsSelfPresenceStamp | null {
	if (!lastWsSelfPresence) {
		return null;
	}
	if (now - lastWsSelfPresence.at > WS_PRESENCE_HTTP_GUARD_MS) {
		return null;
	}
	return lastWsSelfPresence;
}

function resolveManualStatus(
	http: MyPresenceResponse,
	prev?: MyPresenceResponse,
): ManualPresenceStatus {
	if (pendingManualStatus) {
		return pendingManualStatus;
	}
	if (prev && prev.manual_status !== http.manual_status) {
		return prev.manual_status;
	}
	return http.manual_status;
}

/**
 * Apply GET /me/presence without overwriting fresher WS live fields or in-flight optimistic manual.
 */
export function mergeHttpMyPresence(
	http: MyPresenceResponse,
	prev?: MyPresenceResponse,
	now = Date.now(),
): MyPresenceResponse {
	const manual_status = resolveManualStatus(http, prev);
	const ws = wsStampFresh(now);

	if (!ws) {
		return { ...http, manual_status };
	}

	const wsLive = ws.frame.is_online && ws.frame.presence_status !== "offline";
	const httpOffline = !http.is_online || http.effective_status === "offline";

	if (httpOffline && wsLive) {
		return {
			...http,
			manual_status,
			effective_status: ws.frame.presence_status,
			is_online: ws.frame.is_online,
		};
	}

	if (ws.frame.is_online && !http.is_online) {
		return {
			...http,
			manual_status,
			effective_status: ws.frame.presence_status,
			is_online: ws.frame.is_online,
		};
	}

	return { ...http, manual_status };
}

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
	if (!prev?.members.length) {
		return prev;
	}
	const idx = prev.members.findIndex((m) => m.user_id === userId);
	if (idx < 0) {
		return prev;
	}
	const current = prev.members[idx];
	if (!current) {
		return prev;
	}
	const next = patchMemberRow(current, frame);
	if (!next || next === current) {
		return prev;
	}
	const members = [...prev.members];
	members[idx] = next;
	return { ...prev, members };
}

function patchMemberArray(
	prev: GroupMember[] | undefined,
	frame: LivePresenceFrame,
): GroupMember[] | undefined {
	if (!prev?.length) {
		return prev;
	}
	const idx = prev.findIndex((m) => m.user_id === frame.user_id);
	if (idx < 0) {
		return prev;
	}
	const current = prev[idx];
	if (!current) {
		return prev;
	}
	const next = patchMemberRow(current, frame);
	if (!next || next === current) {
		return prev;
	}
	const members = [...prev];
	members[idx] = next;
	return members;
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
	let changed = false;
	queryClient.setQueriesData<GroupMembersPage>(
		{ queryKey: [Q_INVITE_MEMBERS] },
		(prev: GroupMembersPage | undefined) => {
			const next = patchPage(prev, frame.user_id, frame);
			if (next !== prev) {
				changed = true;
			}
			return next;
		},
	);
	queryClient.setQueriesData<GroupMember[]>(
		{ queryKey: [Q_GROUP_ROLE_MEMBERS] },
		(prev: GroupMember[] | undefined) => {
			const next = patchMemberArray(prev, frame);
			if (next !== prev) {
				changed = true;
			}
			return next;
		},
	);
	if (changed && opts?.refreshInvitePreview) {
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
	touchWsSelfPresence(frame);
	queryClient.setQueryData<MyPresenceResponse>([Q_MY_PRESENCE], (prev) => ({
		manual_status: prev?.manual_status ?? pendingManualStatus ?? "online",
		effective_status: frame.presence_status,
		is_online: frame.is_online,
		expires_at: prev?.expires_at ?? "",
	}));
	patchLiveMemberPresence(queryClient, frame);
}

/** Re-apply self row in cached member lists from `Q_MY_PRESENCE` (HTTP refetch / race after WS). */
export function patchSelfMemberPresenceFromMyPresence(
	queryClient: QueryClient,
	selfUserId: string,
): void {
	if (!selfUserId) {
		return;
	}
	const my = queryClient.getQueryData<MyPresenceResponse>([Q_MY_PRESENCE]);
	if (!my) {
		return;
	}
	patchLiveMemberPresence(queryClient, {
		user_id: selfUserId,
		presence_status: my.effective_status,
		is_online: my.is_online,
	});
}
