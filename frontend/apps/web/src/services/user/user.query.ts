import { queryOptions } from "@tanstack/react-query";
import { Q_MY_DEVICES, Q_MY_MFA } from "@/services/user/user.key";
import { getMyDevices, getMyMfaStatus } from "@/services/user/user.service";

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
