import { queryOptions } from "@tanstack/react-query";
import { Q_SESSION, Q_SETUP_STATUS } from "@/services/auth/auth.key";
import { getSession, getSetupStatus } from "@/services/auth/auth.service";

export const setupStatusQueryOpts = () =>
	queryOptions({
		queryKey: [Q_SETUP_STATUS],
		queryFn: getSetupStatus,
		staleTime: 0,
		retry: false,
	});

export const sessionQueryOpts = () =>
	queryOptions({
		queryKey: [Q_SESSION],
		queryFn: getSession,
		staleTime: 60_000,
		retry: false,
	});
