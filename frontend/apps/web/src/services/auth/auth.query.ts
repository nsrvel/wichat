import { queryOptions } from "@tanstack/react-query";
import { Q_INSTANCE_POLICY, Q_SESSION } from "@/services/auth/auth.key";
import { getInstancePolicy, getSession } from "@/services/auth/auth.service";

export const instancePolicyQueryOpts = () =>
	queryOptions({
		queryKey: [Q_INSTANCE_POLICY],
		queryFn: getInstancePolicy,
		staleTime: 60_000,
		retry: false,
	});

export const sessionQueryOpts = () =>
	queryOptions({
		queryKey: [Q_SESSION],
		queryFn: getSession,
		staleTime: 60_000,
		retry: false,
	});
