import { infiniteQueryOptions, queryOptions } from "@tanstack/react-query";
import { mockMyGroupsList } from "@/fixtures/demo-group";
import { groupAuditLogKeys, groupBanKeys, Q_MY_GROUPS } from "@/services/group/group.key";
import { listGroupAuditLogs, listGroupBans, listMyGroups } from "@/services/group/group.service";

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

export const groupBansQueryOpts = (groupId: string, search = "", cursor = "") =>
	queryOptions({
		queryKey: groupBanKeys.list(groupId, search, cursor),
		queryFn: () => listGroupBans(groupId, { search, cursor }),
		staleTime: 30_000,
	});

export const groupAuditLogsInfiniteQueryOpts = (groupId: string, actorId = "", action = "") =>
	infiniteQueryOptions({
		queryKey: groupAuditLogKeys.list(groupId, actorId, action),
		queryFn: ({ pageParam }) =>
			listGroupAuditLogs(groupId, { actorId, action, limit: 30, cursor: pageParam }),
		initialPageParam: "",
		getNextPageParam: (lastPage) => lastPage.next_cursor || undefined,
		staleTime: 15_000,
	});
