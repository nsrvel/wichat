"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { setAccessToken } from "@/integrations/api-client";
import { Q_SESSION, Q_SETUP_STATUS } from "@/services/auth/auth.key";
import { login, logout, setup } from "@/services/auth/auth.service";

export function useLoginMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: login,
		onSuccess: (session) => {
			queryClient.setQueryData([Q_SESSION], session.user);
			queryClient.setQueryData([Q_SETUP_STATUS], { needs_setup: false });
		},
	});
}

export function useSetupMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: setup,
		onSuccess: (session) => {
			queryClient.setQueryData([Q_SESSION], session.user);
			queryClient.setQueryData([Q_SETUP_STATUS], { needs_setup: false });
		},
	});
}

export function useLogoutMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: logout,
		onSettled: () => {
			setAccessToken(null);
			queryClient.removeQueries({ queryKey: [Q_SESSION] });
		},
	});
}
