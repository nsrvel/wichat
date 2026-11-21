"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useParams, usePathname } from "next/navigation";
import { useEffect } from "react";
import { acquireLivePresence, setLivePresenceGroupId } from "@/features/presence/live-connection";
import {
	patchLiveMemberPresence,
	patchSelfMemberPresenceFromMyPresence,
	upsertLiveMyPresence,
} from "@/features/presence/live-cache";
import { normalizePublicGroupId } from "@/features/presence/public-group-id";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import { useQuery } from "@tanstack/react-query";
import { ensureAccessToken } from "@/integrations/api-client";
import { Q_MY_PRESENCE } from "@/services/user/user.key";
import { myPresenceQueryOptions } from "@/services/user/user.query";

/** One tab-wide WS + heartbeats for every authenticated route under AppProviders. */
export function PresenceGate() {
	const queryClient = useQueryClient();
	const { data: user, isSuccess } = useQuery(sessionQueryOpts());
	const pathname = usePathname() ?? "";
	const params = useParams<{ groupId?: string }>();
	const onInviteRoute = pathname.startsWith("/invite/");
	const routeGroupId = normalizePublicGroupId(
		typeof params?.groupId === "string" ? params.groupId : undefined,
	);
	const active = isSuccess && Boolean(user);

	useEffect(() => {
		if (!active) {
			return;
		}
		void ensureAccessToken();
	}, [active]);

	useEffect(() => {
		if (!active) {
			return;
		}
		void queryClient.fetchQuery(myPresenceQueryOptions());
	}, [active, queryClient]);

	useEffect(() => {
		if (!active) {
			return;
		}
		setLivePresenceGroupId(routeGroupId);
	}, [active, routeGroupId]);

	useEffect(() => {
		if (!active || !user?.id) {
			return;
		}
		return acquireLivePresence((frame) => {
			patchLiveMemberPresence(queryClient, frame, { refreshInvitePreview: onInviteRoute });
			upsertLiveMyPresence(queryClient, frame, user.id);
		});
	}, [active, onInviteRoute, user?.id, queryClient]);

	useEffect(() => {
		if (!active || !user?.id) {
			return;
		}
		const selfUserId = user.id;
		return queryClient.getQueryCache().subscribe((event) => {
			if (event.type !== "updated" || event.action.type !== "success") {
				return;
			}
			const root = event.query.queryKey[0];
			if (root === Q_MY_PRESENCE) {
				patchSelfMemberPresenceFromMyPresence(queryClient, selfUserId);
			}
		});
	}, [active, queryClient, user?.id]);

	return null;
}
