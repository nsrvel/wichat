"use client";

import type { GroupMembersPage } from "@wichat/contracts/group/member";
import type { MyPresenceResponse, PresenceStatus } from "@wichat/contracts/user/presence";
import { type QueryClient, type QueryKey, useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import { connectPresenceRealtime } from "@/integrations/realtime/presence-client";
import { Q_INVITE_MEMBERS } from "@/services/invite/invite.query";
import { Q_MY_PRESENCE } from "@/services/user/user.key";

type PresenceSnapshot = {
	is_online: boolean;
	presence_status: PresenceStatus;
};

function applyPresenceToPage(
	prev: GroupMembersPage | undefined,
	userId: string,
	presence: PresenceSnapshot,
): GroupMembersPage | undefined {
	if (!prev) {
		return prev;
	}
	let changed = false;
	const nextMembers = prev.members.map((m) => {
		if (m.user_id !== userId) {
			return m;
		}
		if (m.presence_status === presence.presence_status && m.is_online === presence.is_online) {
			return m;
		}
		changed = true;
		return {
			...m,
			is_online: presence.is_online,
			presence_status: presence.presence_status,
		};
	});
	return changed ? { ...prev, members: nextMembers } : prev;
}

/**
 * Patches all cached group member list pages in-place for a user's presence update.
 */
export function patchMemberListPresence(
	queryClient: QueryClient,
	userId: string,
	presence: PresenceSnapshot,
): void {
	queryClient.setQueriesData<GroupMembersPage>({ queryKey: [Q_INVITE_MEMBERS] }, (prev) =>
		applyPresenceToPage(prev, userId, presence),
	);
}

function patchSingleMemberListQuery(
	queryClient: QueryClient,
	queryKey: QueryKey,
	userId: string,
	presence: PresenceSnapshot,
): void {
	queryClient.setQueryData<GroupMembersPage>(queryKey, (prev) =>
		applyPresenceToPage(prev, userId, presence),
	);
}

/**
 * Keeps the user's presence heartbeat + auto-idle detector active in the workspace
 * and patches cached group member lists in-place when `presence.updated` arrives.
 */
export function usePresenceSync(groupId?: string, enabled = true, currentUserId?: string): void {
	const queryClient = useQueryClient();
	const connRef = useRef<ReturnType<typeof connectPresenceRealtime> | null>(null);
	const groupIdRef = useRef(groupId);
	const currentUserIdRef = useRef(currentUserId);
	const latestSelfPresenceRef = useRef<PresenceSnapshot | null>(null);

	useEffect(() => {
		currentUserIdRef.current = currentUserId;
	}, [currentUserId]);

	useEffect(() => {
		if (!enabled) {
			latestSelfPresenceRef.current = null;
			return;
		}
		const unsubscribeCache = queryClient.getQueryCache().subscribe((cacheEvent) => {
			if (cacheEvent.type !== "updated" || cacheEvent.action.type !== "success") {
				return;
			}
			const selfId = currentUserIdRef.current;
			const selfPresence = latestSelfPresenceRef.current;
			if (!selfId || !selfPresence || cacheEvent.query.queryKey[0] !== Q_INVITE_MEMBERS) {
				return;
			}
			patchSingleMemberListQuery(queryClient, cacheEvent.query.queryKey, selfId, selfPresence);
		});

		const conn = connectPresenceRealtime({
			groupId: groupIdRef.current,
			onPresenceUpdated: (event) => {
				const snapshot: PresenceSnapshot = {
					is_online: event.is_online,
					presence_status: event.presence_status,
				};
				patchMemberListPresence(queryClient, event.user_id, snapshot);
				if (currentUserIdRef.current && event.user_id === currentUserIdRef.current) {
					latestSelfPresenceRef.current = snapshot;
					queryClient.setQueryData<MyPresenceResponse>([Q_MY_PRESENCE], (prev) =>
						prev
							? {
									...prev,
									effective_status: event.presence_status,
									is_online: event.is_online,
								}
							: prev,
					);
				}
			},
		});
		connRef.current = conn;
		return () => {
			unsubscribeCache();
			conn.close();
			connRef.current = null;
		};
	}, [enabled, queryClient]);

	useEffect(() => {
		groupIdRef.current = groupId;
		connRef.current?.setGroupId(groupId);
	}, [groupId]);
}
