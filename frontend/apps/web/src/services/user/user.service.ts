import {
	AccountActionRequestSchema,
	AuthChallengeSchema,
	BeginTotpSetupResponseSchema,
	BeginWebAuthnRegisterResponseSchema,
	CompleteEmailChangeRequestSchema,
	DeviceSessionsResponseSchema,
	FinishWebAuthnRegisterResponseSchema,
	MfaStatusSchema,
	OkSchema,
	RecoveryCodesResponseSchema,
	RequestEmailChangeRequestSchema,
	RevokeDeviceRequestSchema,
	SecurityKeySchema,
	UpdatePasswordRequestSchema,
	UpdateUsernameRequestSchema,
	UpdateUserProfileRequestSchema,
	UserSchema,
	VerifyEmailChangeRequestSchema,
	VerifyEmailChangeResponseSchema,
	type AccountActionRequest,
	type AuthChallenge,
	type BeginTotpSetupResponse,
	type BeginWebAuthnRegisterResponse,
	type CompleteEmailChangeRequest,
	type DeviceSessionsResponse,
	type FinishWebAuthnRegisterResponse,
	type MfaStatus,
	type RecoveryCodesResponse,
	type RequestEmailChangeRequest,
	type RevokeDeviceRequest,
	type SecurityKey,
	type UpdatePasswordRequest,
	type UpdateUsernameRequest,
	type UpdateUserProfileRequest,
	type User,
	type VerifyEmailChangeRequest,
	type VerifyEmailChangeResponse,
} from "@wichat/contracts";
import { api } from "@/integrations/api-client";

export async function uploadMyAvatar(file: File): Promise<User> {
	const form = new FormData();
	form.append("file", file);
	return api.postForm("/users/me/avatar", UserSchema, form);
}

export async function updateMyProfile(body: UpdateUserProfileRequest): Promise<User> {
	const payload = UpdateUserProfileRequestSchema.parse(body);
	return api.patch("/users/me/profile", UserSchema, payload);
}

export type SaveMyProfileInput = UpdateUserProfileRequest & {
	avatarFile?: File | null;
};

export async function saveMyProfile(input: SaveMyProfileInput): Promise<User> {
	const { avatarFile, ...body } = input;
	let user = await updateMyProfile(body);
	if (avatarFile) {
		user = await uploadMyAvatar(avatarFile);
	}
	return user;
}

export async function updateMyUsername(body: UpdateUsernameRequest): Promise<User> {
	const payload = UpdateUsernameRequestSchema.parse(body);
	return api.patch("/users/me/username", UserSchema, payload);
}

export async function requestMyEmailChange(
	body: RequestEmailChangeRequest,
): Promise<AuthChallenge> {
	const payload = RequestEmailChangeRequestSchema.parse(body);
	return api.post("/users/me/email/request", AuthChallengeSchema, payload);
}

export async function verifyMyEmailChange(
	body: VerifyEmailChangeRequest,
): Promise<VerifyEmailChangeResponse> {
	const payload = VerifyEmailChangeRequestSchema.parse(body);
	return api.post("/users/me/email/verify", VerifyEmailChangeResponseSchema, payload);
}

export async function completeMyEmailChange(body: CompleteEmailChangeRequest): Promise<User> {
	const payload = CompleteEmailChangeRequestSchema.parse(body);
	return api.patch("/users/me/email", UserSchema, payload);
}

export async function updateMyPassword(body: UpdatePasswordRequest): Promise<{ ok: boolean }> {
	const payload = UpdatePasswordRequestSchema.parse(body);
	return api.patch("/users/me/password", OkSchema, payload);
}

export async function getMyMfaStatus(): Promise<MfaStatus> {
	return api.get("/users/me/mfa", MfaStatusSchema);
}

export async function beginTotpSetup(
	current_password: string,
): Promise<BeginTotpSetupResponse> {
	return api.post("/users/me/mfa/totp/begin", BeginTotpSetupResponseSchema, {
		current_password,
	});
}

export async function confirmTotpSetup(body: {
	setup_token: string;
	code: string;
}): Promise<RecoveryCodesResponse> {
	return api.post("/users/me/mfa/totp/confirm", RecoveryCodesResponseSchema, body);
}

export async function disableTotp(
	current_password: string,
): Promise<{ ok: boolean }> {
	return api.delete("/users/me/mfa/totp", OkSchema, { current_password });
}

export async function beginWebAuthnRegister(body: {
	current_password: string;
	name: string;
}): Promise<BeginWebAuthnRegisterResponse> {
	return api.post(
		"/users/me/mfa/security-keys/begin",
		BeginWebAuthnRegisterResponseSchema,
		body,
	);
}

export async function finishWebAuthnRegister(body: {
	challenge_id: string;
	credential: unknown;
}): Promise<FinishWebAuthnRegisterResponse> {
	return api.post(
		"/users/me/mfa/security-keys/finish",
		FinishWebAuthnRegisterResponseSchema,
		body,
	);
}

export async function renameSecurityKey(body: {
	id: string;
	name: string;
}): Promise<SecurityKey> {
	return api.patch(
		`/users/me/mfa/security-keys/${encodeURIComponent(body.id)}`,
		SecurityKeySchema,
		{ name: body.name },
	);
}

export async function deleteSecurityKey(body: {
	id: string;
	current_password: string;
}): Promise<{ ok: boolean }> {
	return api.delete(
		`/users/me/mfa/security-keys/${encodeURIComponent(body.id)}`,
		OkSchema,
		{ current_password: body.current_password },
	);
}

export async function regenerateRecoveryCodes(
	current_password: string,
): Promise<RecoveryCodesResponse> {
	return api.post(
		"/users/me/mfa/recovery-codes/regenerate",
		RecoveryCodesResponseSchema,
		{ current_password },
	);
}

export async function disableMyAccount(body: AccountActionRequest): Promise<{ ok: boolean }> {
	const payload = AccountActionRequestSchema.parse(body);
	return api.post("/users/me/disable", OkSchema, payload);
}

export async function deleteMyAccount(body: AccountActionRequest): Promise<{ ok: boolean }> {
	const payload = AccountActionRequestSchema.parse(body);
	return api.post("/users/me/delete", OkSchema, payload);
}

export async function getMyDevices(): Promise<DeviceSessionsResponse> {
	return api.get("/users/me/devices", DeviceSessionsResponseSchema);
}

export async function revokeMyDevice(body: {
	id: string;
	current_password?: string;
}): Promise<{ ok: boolean }> {
	const payload: RevokeDeviceRequest = RevokeDeviceRequestSchema.parse({
		current_password: body.current_password ?? "",
	});
	return api.delete(`/users/me/devices/${encodeURIComponent(body.id)}`, OkSchema, payload);
}

export async function revokeOtherDevices(body: RevokeDeviceRequest): Promise<{ ok: boolean }> {
	const payload = RevokeDeviceRequestSchema.parse(body);
	return api.post("/users/me/devices/revoke-others", OkSchema, payload);
}

