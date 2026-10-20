"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Q_JOINABLE_GROUPS, Q_MY_GROUPS } from "@/services/group/group.key";
import { createGroup, joinGroup } from "@/services/group/group.service";

export function useCreateGroupMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: createGroup,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
			void queryClient.invalidateQueries({ queryKey: [Q_JOINABLE_GROUPS] });
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
