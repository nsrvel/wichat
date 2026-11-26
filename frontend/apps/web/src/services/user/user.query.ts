import type { QueryClient } from "@tanstack/react-query";
import { queryOptions } from "@tanstack/react-query";
import { mergeHttpMyPresence } from "@/features/presence/live-cache";
import {
	Q_MY_DEVICES,
	Q_MY_MFA,
	Q_MY_PRESENCE,
	Q_USER_RELATIONSHIPS,
} from "@/services/user/user.key";
import {
	getMyDevices,
	getMyMfaStatus,
	getMyPresence,
	listUserRelationships,
} from "@/services/user/user.service";

export function myMfaStatusQueryOptions(enabled = true) {
	return queryOptions({
		queryKey: [Q_MY_MFA],
		queryFn: getMyMfaStatus,
		enabled,
	});
}

export function myDevicesQueryOptions() {
	return queryOptions({
		queryKey: [Q_MY_DEVICES],
		queryFn: getMyDevices,
	});
}

export function myPresenceQueryOptions(queryClient?: QueryClient) {
	return queryOptions({
		queryKey: [Q_MY_PRESENCE],
		queryFn: async () => {
			const http = await getMyPresence();
			const prev = queryClient?.getQueryData<typeof http>([Q_MY_PRESENCE]);
			return mergeHttpMyPresence(http, prev);
		},
		staleTime: Number.POSITIVE_INFINITY,
		gcTime: 30 * 60_000,
		refetchOnWindowFocus: false,
		refetchOnReconnect: false,
	});
}

export function userRelationshipsQueryOptions() {
	return queryOptions({
		queryKey: [Q_USER_RELATIONSHIPS],
		queryFn: async () => {
			const data = await listUserRelationships();
			return data.relationships;
		},
		staleTime: 60_000,
	});
}
