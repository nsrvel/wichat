"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Q_GROUP_CHANNELS } from "@/services/channel/channel.key";
import { Q_JOINABLE_GROUPS, Q_MY_GROUPS } from "@/services/group/group.key";
import type { Group } from "@wichat/contracts";
import {
	createGroup,
	deleteGroup,
	joinGroup,
	leaveGroup,
	reorderMyGroups,
} from "@/services/group/group.service";

export function useCreateGroupMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: createGroup,
		onSuccess: (group) => {
			const owned: Group = { ...group, role: group.role ?? "owner" };
			queryClient.setQueryData<Group[]>([Q_MY_GROUPS], (prev) => [
				...(prev ?? []).filter((g) => g.id !== owned.id),
				owned,
			]);
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
			void queryClient.invalidateQueries({ queryKey: [Q_JOINABLE_GROUPS] });
			void queryClient.invalidateQueries({ queryKey: [Q_GROUP_CHANNELS, group.id] });
		},
	});
}

export function useJoinGroupMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: joinGroup,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
			void queryClient.invalidateQueries({ queryKey: [Q_JOINABLE_GROUPS] });
		},
	});
}

export function useLeaveGroupMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: leaveGroup,
		onSuccess: (_data, groupId) => {
			queryClient.setQueryData<Group[]>([Q_MY_GROUPS], (prev) =>
				(prev ?? []).filter((g) => g.id !== groupId),
			);
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
			void queryClient.invalidateQueries({ queryKey: [Q_JOINABLE_GROUPS] });
		},
	});
}

export function useReorderMyGroupsMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: reorderMyGroups,
		onMutate: async (groupIds) => {
			await queryClient.cancelQueries({ queryKey: [Q_MY_GROUPS] });
			const previous = queryClient.getQueryData<Group[]>([Q_MY_GROUPS]);
			if (previous) {
				const byId = new Map(previous.map((g) => [g.id, g]));
				const reordered = groupIds
					.map((id) => byId.get(id))
					.filter((g): g is Group => g != null);
				if (reordered.length === previous.length) {
					queryClient.setQueryData<Group[]>([Q_MY_GROUPS], reordered);
				}
			}
			return { previous };
		},
		onError: (_err, _ids, ctx) => {
			if (ctx?.previous) {
				queryClient.setQueryData([Q_MY_GROUPS], ctx.previous);
			}
		},
		onSettled: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
		},
	});
}

export function useDeleteGroupMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: deleteGroup,
		onSuccess: (_data, groupId) => {
			queryClient.setQueryData<Group[]>([Q_MY_GROUPS], (prev) =>
				(prev ?? []).filter((g) => g.id !== groupId),
			);
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
			void queryClient.invalidateQueries({ queryKey: [Q_JOINABLE_GROUPS] });
		},
	});
}
