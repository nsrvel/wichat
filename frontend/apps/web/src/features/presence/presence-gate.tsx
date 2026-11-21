"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useParams, usePathname } from "next/navigation";
import { useEffect } from "react";
import { acquireLivePresence, setLivePresenceGroupId } from "@/features/presence/live-connection";
import {
	patchLiveMemberPresence,
	upsertLiveMyPresence,
} from "@/features/presence/live-cache";
import { normalizePublicGroupId } from "@/features/presence/public-group-id";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import { useQuery } from "@tanstack/react-query";
import { ensureAccessToken } from "@/integrations/api-client";

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

	return null;
}