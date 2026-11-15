import { queryOptions } from "@tanstack/react-query";
import { Q_MY_DEVICES, Q_MY_MFA, Q_MY_PRESENCE } from "@/services/user/user.key";
import { getMyDevices, getMyMfaStatus, getMyPresence } from "@/services/user/user.service";

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
	});
}
