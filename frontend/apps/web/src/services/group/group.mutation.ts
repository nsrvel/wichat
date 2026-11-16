"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Q_GROUP_CHANNELS } from "@/services/channel/channel.key";
import { groupBanKeys, Q_MY_GROUPS } from "@/services/group/group.key";
import { groupRoleKeys } from "@/services/group-role/group-role.key";
import { Q_INVITE_MEMBERS } from "@/services/invite/invite.query";
import type {
	BanGroupMembersRequest,
	TimeoutGroupMemberRequest,
} from "@wichat/contracts/group/member";
import type { Group } from "@wichat/contracts/group";
import {
	banGroupMembers,
	createGroup,
	deleteGroup,
	joinGroup,
	kickGroupMember,
	leaveGroup,
	reorderMyGroups,
	saveGroupProfile,
	timeoutGroupMember,
	unbanGroupMember,
	type SaveGroupProfileInput,
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
				const reordered = groupIds.map((id) => byId.get(id)).filter((g): g is Group => g != null);
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

export function useSaveGroupProfileMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({ groupId, input }: { groupId: string; input: SaveGroupProfileInput }) =>
			saveGroupProfile(groupId, input),
		onSuccess: (group) => {
			queryClient.setQueryData<Group[]>([Q_MY_GROUPS], (prev) =>
				(prev ?? []).map((g) =>
					g.id === group.id ? { ...g, ...group, role: g.role ?? group.role } : g,
				),
			);
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
		},
	});
}

export function useKickGroupMemberMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({ userId, reason }: { userId: string; reason?: string }) =>
			kickGroupMember(groupId, userId, reason),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_INVITE_MEMBERS, groupId] });
			void queryClient.invalidateQueries({ queryKey: groupRoleKeys.list(groupId) });
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
		},
	});
}

export function useBanGroupMembersMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (payload: BanGroupMembersRequest) => banGroupMembers(groupId, payload),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_INVITE_MEMBERS, groupId] });
			void queryClient.invalidateQueries({ queryKey: groupRoleKeys.list(groupId) });
			void queryClient.invalidateQueries({ queryKey: groupBanKeys.all(groupId) });
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
		},
	});
}

export function useUnbanGroupMemberMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (userId: string) => unbanGroupMember(groupId, userId),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: groupBanKeys.all(groupId) });
		},
	});
}

export function useTimeoutGroupMemberMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({ userId, ...payload }: TimeoutGroupMemberRequest & { userId: string }) =>
			timeoutGroupMember(groupId, userId, payload),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_INVITE_MEMBERS, groupId] });
		},
	});
}
