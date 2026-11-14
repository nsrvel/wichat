import { useMutation, useQueryClient } from "@tanstack/react-query";
import { setAccessToken } from "@/integrations/api-client";
import { Q_SESSION } from "@/services/auth/auth.key";
import { Q_MY_DEVICES, Q_MY_MFA } from "@/services/user/user.key";
import {
	beginTotpSetup,
	beginWebAuthnRegister,
	completeMyEmailChange,
	confirmTotpSetup,
	deleteMyAccount,
	deleteSecurityKey,
	disableMyAccount,
	disableTotp,
	finishWebAuthnRegister,
	regenerateRecoveryCodes,
	renameSecurityKey,
	requestMyEmailChange,
	revokeMyDevice,
	revokeOtherDevices,
	saveMyProfile,
	updateMyPassword,
	updateMyUsername,
	verifyMyEmailChange,
	type SaveMyProfileInput,
} from "@/services/user/user.service";

export function useSaveMyProfileMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (input: SaveMyProfileInput) => saveMyProfile(input),
		onSuccess: (user) => {
			queryClient.setQueryData([Q_SESSION], user);
			void queryClient.invalidateQueries({ queryKey: [Q_SESSION] });
		},
	});
}

export function useUpdateMyUsernameMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: updateMyUsername,
		onSuccess: (user) => {
			queryClient.setQueryData([Q_SESSION], user);
			void queryClient.invalidateQueries({ queryKey: [Q_SESSION] });
		},
	});
}

export function useRequestMyEmailChangeMutation() {
	return useMutation({
		mutationFn: requestMyEmailChange,
	});
}

export function useVerifyMyEmailChangeMutation() {
	return useMutation({
		mutationFn: verifyMyEmailChange,
	});
}

export function useCompleteMyEmailChangeMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: completeMyEmailChange,
		onSuccess: (user) => {
			queryClient.setQueryData([Q_SESSION], user);
			void queryClient.invalidateQueries({ queryKey: [Q_SESSION] });
		},
	});
}

export function useUpdateMyPasswordMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: updateMyPassword,
		onSuccess: () => {
			queryClient.setQueryData<Record<string, unknown> | undefined>([Q_SESSION], (prev) =>
				prev ? { ...prev, has_password: true } : prev,
			);
			void queryClient.invalidateQueries({ queryKey: [Q_SESSION] });
		},
	});
}

export function useBeginTotpSetupMutation() {
	return useMutation({ mutationFn: beginTotpSetup });
}

export function useConfirmTotpSetupMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: confirmTotpSetup,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_MY_MFA] });
		},
	});
}

export function useDisableTotpMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: disableTotp,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_MY_MFA] });
		},
	});
}

export function useBeginWebAuthnRegisterMutation() {
	return useMutation({ mutationFn: beginWebAuthnRegister });
}

export function useFinishWebAuthnRegisterMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: finishWebAuthnRegister,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_MY_MFA] });
		},
	});
}

export function useRenameSecurityKeyMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: renameSecurityKey,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_MY_MFA] });
		},
	});
}

export function useDeleteSecurityKeyMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: deleteSecurityKey,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_MY_MFA] });
		},
	});
}

export function useRegenerateRecoveryCodesMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: regenerateRecoveryCodes,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_MY_MFA] });
		},
	});
}

export function useDisableMyAccountMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: disableMyAccount,
		onSuccess: () => {
			setAccessToken(null);
			queryClient.clear();
		},
	});
}

export function useDeleteMyAccountMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: deleteMyAccount,
		onSuccess: () => {
			setAccessToken(null);
			queryClient.clear();
		},
	});
}

export function useRevokeMyDeviceMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: revokeMyDevice,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_MY_DEVICES] });
		},
	});
}

export function useRevokeOtherDevicesMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: revokeOtherDevices,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_MY_DEVICES] });
		},
	});
}

