"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { GroupRolesListResponse, UpdateGroupRoleRequest } from "@wichat/contracts/group/role";
import { Q_MY_GROUPS } from "@/services/group/group.key";
import { groupRoleKeys } from "@/services/group-role/group-role.key";
import {
	addGroupRoleMembers,
	createGroupRole,
	deleteGroupRole,
	duplicateGroupRole,
	removeGroupRoleMember,
	updateGroupHideRoles,
	updateGroupRole,
} from "@/services/group-role/group-role.service";
import { Q_INVITE_MEMBERS } from "@/services/invite/invite.query";

export function useCreateGroupRoleMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (name?: string) => createGroupRole(groupId, { name: name ?? "new role" }),
		onSuccess: (created) => {
			queryClient.setQueryData<GroupRolesListResponse>(groupRoleKeys.list(groupId), (prev) =>
				prev ? { ...prev, roles: [...prev.roles, created] } : prev,
			);
			void queryClient.invalidateQueries({ queryKey: groupRoleKeys.list(groupId) });
		},
	});
}

export function useUpdateGroupRoleMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({ roleId, payload }: { roleId: string; payload: UpdateGroupRoleRequest }) =>
			updateGroupRole(groupId, roleId, payload),
		onSuccess: (updated) => {
			queryClient.setQueryData<GroupRolesListResponse>(groupRoleKeys.list(groupId), (prev) =>
				prev
					? {
							...prev,
							roles: prev.roles.map((r) => (r.id === updated.id ? updated : r)),
						}
					: prev,
			);
			void queryClient.invalidateQueries({ queryKey: groupRoleKeys.list(groupId) });
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
			void queryClient.invalidateQueries({ queryKey: [Q_INVITE_MEMBERS, groupId] });
		},
	});
}

export function useDuplicateGroupRoleMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (roleId: string) => duplicateGroupRole(groupId, roleId),
		onSuccess: (created) => {
			queryClient.setQueryData<GroupRolesListResponse>(groupRoleKeys.list(groupId), (prev) =>
				prev ? { ...prev, roles: [...prev.roles, created] } : prev,
			);
			void queryClient.invalidateQueries({ queryKey: groupRoleKeys.list(groupId) });
		},
	});
}

export function useDeleteGroupRoleMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (roleId: string) => deleteGroupRole(groupId, roleId),
		onSuccess: (_res, roleId) => {
			queryClient.setQueryData<GroupRolesListResponse>(groupRoleKeys.list(groupId), (prev) =>
				prev ? { ...prev, roles: prev.roles.filter((r) => r.id !== roleId) } : prev,
			);
			void queryClient.invalidateQueries({ queryKey: groupRoleKeys.list(groupId) });
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
			void queryClient.invalidateQueries({ queryKey: [Q_INVITE_MEMBERS, groupId] });
		},
	});
}

export function useAddGroupRoleMembersMutation(groupId: string, roleId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (userIds: string[]) => addGroupRoleMembers(groupId, roleId, { user_ids: userIds }),
		onSuccess: (updated) => {
			queryClient.setQueryData<GroupRolesListResponse>(groupRoleKeys.list(groupId), (prev) =>
				prev
					? {
							...prev,
							roles: prev.roles.map((r) => (r.id === updated.id ? updated : r)),
						}
					: prev,
			);
			void queryClient.invalidateQueries({ queryKey: groupRoleKeys.members(groupId, roleId) });
			void queryClient.invalidateQueries({ queryKey: groupRoleKeys.list(groupId) });
			void queryClient.invalidateQueries({ queryKey: [Q_INVITE_MEMBERS, groupId] });
		},
	});
}

export function useRemoveGroupRoleMemberMutation(groupId: string, roleId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (userId: string) => removeGroupRoleMember(groupId, roleId, userId),
		onSuccess: (updated) => {
			queryClient.setQueryData<GroupRolesListResponse>(groupRoleKeys.list(groupId), (prev) =>
				prev
					? {
							...prev,
							roles: prev.roles.map((r) => (r.id === updated.id ? updated : r)),
						}
					: prev,
			);
			void queryClient.invalidateQueries({ queryKey: groupRoleKeys.members(groupId, roleId) });
			void queryClient.invalidateQueries({ queryKey: groupRoleKeys.list(groupId) });
			void queryClient.invalidateQueries({ queryKey: [Q_INVITE_MEMBERS, groupId] });
		},
	});
}

export function useUpdateGroupHideRolesMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (hideRoles: boolean) => updateGroupHideRoles(groupId, hideRoles),
		onMutate: async (nextHide) => {
			await queryClient.cancelQueries({ queryKey: groupRoleKeys.list(groupId) });
			const previous = queryClient.getQueryData<GroupRolesListResponse>(
				groupRoleKeys.list(groupId),
			);
			queryClient.setQueryData<GroupRolesListResponse>(groupRoleKeys.list(groupId), (prev) =>
				prev ? { ...prev, hide_roles: nextHide } : prev,
			);
			return { previous };
		},
		onError: (_err, _next, ctx) => {
			if (ctx?.previous) {
				queryClient.setQueryData(groupRoleKeys.list(groupId), ctx.previous);
			}
		},
		onSettled: () => {
			void queryClient.invalidateQueries({ queryKey: groupRoleKeys.list(groupId) });
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
			void queryClient.invalidateQueries({ queryKey: [Q_INVITE_MEMBERS, groupId] });
		},
	});
}
