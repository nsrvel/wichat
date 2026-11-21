"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo } from "react";
import { onlineCountFromMemberCache } from "@/features/presence/live-cache";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";

/** Live online tally from member-list cache (HTTP seed + WS patches). */
export function useGroupOnlineCount(groupId: string | undefined): number {
	const queryClient = useQueryClient();
	const membersQuery = useQuery({
		...groupMembersQueryOpts(groupId ?? "", ""),
		enabled: Boolean(groupId),
	});

	return useMemo(() => {
		if (!groupId) {
			return 0;
		}
		const fromCache = onlineCountFromMemberCache(queryClient, groupId);
		if (fromCache > 0) {
			return fromCache;
		}
		return membersQuery.data?.members.filter((m) => m.is_online).length ?? 0;
	}, [groupId, queryClient, membersQuery.data]);
}
