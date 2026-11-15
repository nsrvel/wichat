"use client";

import type { GroupMembersPage } from "@wichat/contracts/group/member";
import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import { connectPresenceRealtime } from "@/integrations/realtime/presence-client";
import { Q_INVITE_MEMBERS } from "@/services/invite/invite.query";

/**
 * Keeps the user's presence heartbeat + auto-idle detector active in the workspace
 * and patches cached group member lists in-place when `presence.updated` arrives.
 */
export function usePresenceSync(groupId?: string, enabled = true): void {
	const queryClient = useQueryClient();
	const connRef = useRef<ReturnType<typeof connectPresenceRealtime> | null>(null);
	const groupIdRef = useRef(groupId);

	useEffect(() => {
		if (!enabled) {
			return;
		}
		const conn = connectPresenceRealtime({
			groupId: groupIdRef.current,
			onPresenceUpdated: (event) => {
				queryClient.setQueriesData<GroupMembersPage>(
					{ queryKey: [Q_INVITE_MEMBERS] },
					(prev) => {
						if (!prev) {
							return prev;
						}
						let changed = false;
						const nextMembers = prev.members.map((m) => {
							if (m.user_id !== event.user_id) {
								return m;
							}
							if (
								m.presence_status === event.presence_status &&
								m.is_online === event.is_online
							) {
								return m;
							}
							changed = true;
							return {
								...m,
								is_online: event.is_online,
								presence_status: event.presence_status,
							};
						});
						return changed ? { ...prev, members: nextMembers } : prev;
					},
				);
			},
		});
		connRef.current = conn;
		return () => {
			conn.close();
			connRef.current = null;
		};
	}, [enabled, queryClient]);

	useEffect(() => {
		groupIdRef.current = groupId;
		connRef.current?.setGroupId(groupId);
	}, [groupId]);
}
