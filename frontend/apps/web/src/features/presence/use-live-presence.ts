"use client";

import { useQuery } from "@tanstack/react-query";
import { presenceDotFrom } from "@/features/presence/presence-dot";
import type { UserPresenceStatus } from "@/lib/user-status";
import { myPresenceQueryOptions } from "@/services/user/user.query";

/**
 * Self presence for UI dots (effective via WS) + manual status for the picker only.
 */
export function useLivePresence(enabled = true) {
	const query = useQuery({
		...myPresenceQueryOptions(),
		enabled,
	});

	const dot = presenceDotFrom(query.data);
	const manualStatus: UserPresenceStatus = query.data?.manual_status ?? "online";

	return {
		query,
		dot,
		manualStatus,
		avatarDot: dot ?? (query.isPending ? undefined : ("offline" as const)),
	};
}
