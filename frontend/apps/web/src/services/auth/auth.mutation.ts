"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { resetWorkspaceShellReady } from "@/features/shell/workspace-ready-gate";
import { useRouter } from "next/navigation";
import { setAccessToken } from "@/integrations/api-client";
import { Q_SESSION } from "@/services/auth/auth.key";
import { applySession } from "@/services/auth/apply-session";
import { Q_APP_BOOTSTRAP } from "@/services/auth/bootstrap-app";
import { Q_MY_GROUPS } from "@/services/group/group.key";
import {
	beginLoginWebAuthn,
	forgotPasswordRequest,
	forgotPasswordReset,
	forgotPasswordVerify,
	login,
	logout,
	register,
	resendLoginCode,
	resendRegisterCode,
	verifyLogin,
	verifyLoginMfa,
	verifyRegister,
} from "@/services/auth/auth.service";

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

export function useBeginLoginWebAuthnMutation() {
	return useMutation({ mutationFn: beginLoginWebAuthn });
}

export function useVerifyLoginMfaMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: verifyLoginMfa,
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
	const router = useRouter();
	return useMutation({
		mutationFn: logout,
		onSettled: () => {
			resetWorkspaceShellReady();
			setAccessToken(null);
			queryClient.setQueryData([Q_APP_BOOTSTRAP], null);
			queryClient.removeQueries({ queryKey: [Q_SESSION] });
			queryClient.removeQueries({ queryKey: [Q_MY_GROUPS] });
			router.replace("/sign-in");
		},
	});
}
