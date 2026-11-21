import { useMutation, useQueryClient } from "@tanstack/react-query";
import { patchLiveMemberPresence } from "@/features/presence/live-cache";
import { setAccessToken } from "@/integrations/api-client";
import { Q_SESSION } from "@/services/auth/auth.key";
import {
	Q_MY_DEVICES,
	Q_MY_MFA,
	Q_MY_PRESENCE,
	Q_USER_RELATIONSHIPS,
} from "@/services/user/user.key";
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
	setUserRelationship,
	updateMyPassword,
	updateMyPreferences,
	updateMyPresence,
	updateMyStatusNote,
	updateMyUsername,
	verifyMyEmailChange,
	type SaveMyProfileInput,
} from "@/services/user/user.service";
import type {
	UpdateUserPreferencesRequest,
	UpdateUserStatusNoteRequest,
	User,
} from "@wichat/contracts/user";
import type { MyPresenceResponse } from "@wichat/contracts/user/presence";

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

export function useUpdateMyStatusNoteMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (body: UpdateUserStatusNoteRequest) => updateMyStatusNote(body),
		onMutate: async (body) => {
			await queryClient.cancelQueries({ queryKey: [Q_SESSION] });
			const previous = queryClient.getQueryData<User>([Q_SESSION]);
			if (previous) {
				queryClient.setQueryData<User>([Q_SESSION], {
					...previous,
					status_note_text: body.text ?? "",
					status_note_emoji: body.emoji ?? "",
					status_note_expires_at: body.expires_at ?? null,
				});
			}
			return { previous };
		},
		onError: (_err, _body, context) => {
			if (context?.previous) {
				queryClient.setQueryData([Q_SESSION], context.previous);
			}
		},
		onSuccess: (user) => {
			queryClient.setQueryData([Q_SESSION], user);
		},
		onSettled: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_SESSION] });
		},
	});
}

export function useUpdateMyPreferencesMutation() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (body: UpdateUserPreferencesRequest) => updateMyPreferences(body),
		onMutate: async (body) => {
			await queryClient.cancelQueries({ queryKey: [Q_SESSION] });
			const previous = queryClient.getQueryData<User>([Q_SESSION]);
			if (previous) {
				queryClient.setQueryData<User>([Q_SESSION], {
					...previous,
					locale: body.locale,
					time_format: body.time_format,
				});
			}
			return { previous };
		},
		onError: (_err, _body, context) => {
			if (context?.previous) {
				queryClient.setQueryData([Q_SESSION], context.previous);
			}
		},
		onSuccess: (user) => {
			queryClient.setQueryData([Q_SESSION], user);
		},
		onSettled: () => {
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

export function useUpdateMyPresenceMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: updateMyPresence,
		onMutate: async (input) => {
			await queryClient.cancelQueries({ queryKey: [Q_MY_PRESENCE] });
			const prev = queryClient.getQueryData<MyPresenceResponse>([Q_MY_PRESENCE]);
			if (prev) {
				queryClient.setQueryData<MyPresenceResponse>([Q_MY_PRESENCE], {
					...prev,
					manual_status: input.status,
				});
			}
			return { prev };
		},
		onError: (_err, _input, ctx) => {
			if (ctx?.prev) {
				queryClient.setQueryData([Q_MY_PRESENCE], ctx.prev);
			}
		},
		onSuccess: (res) => {
			queryClient.setQueryData([Q_MY_PRESENCE], res);
			const sessionUser = queryClient.getQueryData<User>([Q_SESSION]);
			if (sessionUser?.id) {
				patchLiveMemberPresence(queryClient, {
					user_id: sessionUser.id,
					is_online: res.is_online,
					presence_status: res.effective_status,
				});
			}
		},
	});
}

export function useSetUserRelationshipMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: setUserRelationship,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_USER_RELATIONSHIPS] });
		},
	});
}
