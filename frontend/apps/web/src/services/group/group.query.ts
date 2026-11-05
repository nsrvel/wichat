import { queryOptions } from "@tanstack/react-query";
import { mockMyGroupsList } from "@/fixtures/demo-group";
import { Q_JOINABLE_GROUPS, Q_MY_GROUPS } from "@/services/group/group.key";
import { listJoinableGroups, listMyGroups } from "@/services/group/group.service";

const UI_SAMPLE_GROUPS_KEY = "ui-sample";

export const myGroupsQueryOpts = () =>
	queryOptions({
		queryKey: [Q_MY_GROUPS],
		queryFn: async () => {
			const data = await listMyGroups();
			return data.groups;
		},
		staleTime: 5 * 60_000,
		gcTime: 30 * 60_000,
		refetchOnWindowFocus: false,
	});

/** Fixture workspace for `/dev/chat-ui` — does not call the API. */
export const myGroupsUiSampleQueryOpts = () =>
	queryOptions({
		queryKey: [Q_MY_GROUPS, UI_SAMPLE_GROUPS_KEY],
		queryFn: async () => mockMyGroupsList.groups,
		staleTime: Number.POSITIVE_INFINITY,
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
