import { queryOptions } from "@tanstack/react-query";
import { Q_JOINABLE_GROUPS, Q_MY_GROUPS } from "@/services/group/group.key";
import { listJoinableGroups, listMyGroups } from "@/services/group/group.service";

export const myGroupsQueryOpts = () =>
	queryOptions({
		queryKey: [Q_MY_GROUPS],
		queryFn: async () => {
			const data = await listMyGroups();
			return data.groups;
		},
		staleTime: 30_000,
	});

export const joinableGroupsQueryOpts = () =>
	queryOptions({
		queryKey: [Q_JOINABLE_GROUPS],
		queryFn: async () => {
			const data = await listJoinableGroups();
			return data.groups;
		},
		staleTime: 15_000,
	});
