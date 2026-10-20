"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { setAccessToken } from "@/integrations/api-client";
import { Q_INSTANCE_POLICY, Q_SESSION } from "@/services/auth/auth.key";
import { Q_JOINABLE_GROUPS, Q_MY_GROUPS } from "@/services/group/group.key";
import {
	forgotPasswordRequest,
	forgotPasswordReset,
	forgotPasswordVerify,
	login,
	logout,
	register,
	registerGuest,
	resendLoginCode,
	resendRegisterCode,
	verifyLogin,
	verifyRegister,
} from "@/services/auth/auth.service";

function applySession(
	queryClient: ReturnType<typeof useQueryClient>,
	session: { user: { id: string } },
) {
	queryClient.setQueryData([Q_SESSION], session.user);
	void queryClient.invalidateQueries({ queryKey: [Q_INSTANCE_POLICY] });
}

export function useLoginMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: login,
		onSuccess: (step) => {
			if (step.kind === "session") {
				applySession(queryClient, step.session);
			}
		},
	});
}

export function useVerifyLoginMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: verifyLogin,
		onSuccess: (session) => applySession(queryClient, session),
	});
}

export function useResendLoginCodeMutation() {
	return useMutation({ mutationFn: resendLoginCode });
}

export function useRegisterMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: register,
		onSuccess: (step) => {
			if (step.kind === "session") {
				applySession(queryClient, step.session);
			}
		},
	});
}

export function useVerifyRegisterMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: verifyRegister,
		onSuccess: (session) => applySession(queryClient, session),
	});
}

export function useResendRegisterCodeMutation() {
	return useMutation({ mutationFn: resendRegisterCode });
}

export function useRegisterGuestMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: registerGuest,
		onSuccess: (session) => {
			queryClient.setQueryData([Q_SESSION], session.user);
			void queryClient.invalidateQueries({ queryKey: [Q_INSTANCE_POLICY] });
		},
	});
}

export function useForgotPasswordRequestMutation() {
	return useMutation({ mutationFn: forgotPasswordRequest });
}

export function useForgotPasswordVerifyMutation() {
	return useMutation({ mutationFn: forgotPasswordVerify });
}

export function useForgotPasswordResetMutation() {
	return useMutation({ mutationFn: forgotPasswordReset });
}

export function useLogoutMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: logout,
		onSettled: () => {
			setAccessToken(null);
			queryClient.removeQueries({ queryKey: [Q_SESSION] });
			queryClient.removeQueries({ queryKey: [Q_MY_GROUPS] });
			queryClient.removeQueries({ queryKey: [Q_JOINABLE_GROUPS] });
		},
	});
}
