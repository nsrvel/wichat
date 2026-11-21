import { queryOptions } from "@tanstack/react-query";
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

export function myMfaStatusQueryOptions() {
	return queryOptions({
		queryKey: [Q_MY_MFA],
		queryFn: getMyMfaStatus,
	});
}

export function myDevicesQueryOptions() {
	return queryOptions({
		queryKey: [Q_MY_DEVICES],
		queryFn: getMyDevices,
	});
}

export function myPresenceQueryOptions() {
	return queryOptions({
		queryKey: [Q_MY_PRESENCE],
		queryFn: getMyPresence,
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
