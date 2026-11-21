import {
	AuthChallengeSchema,
	CheckUsernameResponseSchema,
	InstancePolicySchema,
	parsePasswordAuthStep,
	ResendEmailCodeResponseSchema,
	SessionSchema,
	type LoginRequest,
	type PasswordAuthStep,
	type RegisterRequest,
	type ResendEmailCode,
	type VerifyEmailCode,
} from "@wichat/contracts/auth";
import {
	BeginLoginWebAuthnResponseSchema,
	VerifyLoginMfaRequestSchema,
	type BeginLoginWebAuthnResponse,
	type VerifyLoginMfaRequest,
} from "@wichat/contracts/auth/mfa";
import {
	ForgotPasswordRequestResponseSchema,
	ForgotPasswordVerifyResponseSchema,
	type ForgotPasswordRequest,
	type ForgotPasswordReset,
	type ForgotPasswordVerify,
} from "@wichat/contracts/auth/recovery";
import { OkSchema } from "@wichat/contracts/common";
import { UserSchema, type User } from "@wichat/contracts/user";
import { z } from "zod";
import { ApiError } from "@wichat/api-client/api-error";
import { env } from "@/lib/env";
import { api, ensureAccessToken, setAccessToken } from "@/integrations/api-client";

export function googleStartUrl(returnTo?: string, turnstileToken?: string) {
	const url = new URL("/api/v1/auth/google", env.NEXT_PUBLIC_API_URL);
	const next = returnTo?.trim();
	if (next?.startsWith("/") && !next.startsWith("//")) {
		url.searchParams.set("next", next);
	}
	const token = turnstileToken?.trim();
	if (token) {
		url.searchParams.set("turnstile_token", token);
	}
	return url.toString();
}

export async function getInstancePolicy() {
	return api.get("/auth/setup-status", InstancePolicySchema);
}

export async function checkUsername(username: string) {
	const params = new URLSearchParams({ username });
	return api.get(`/users/check-username?${params.toString()}`, CheckUsernameResponseSchema);
}

const authStepDataSchema = z.union([SessionSchema, AuthChallengeSchema]);

export async function login(payload: LoginRequest): Promise<PasswordAuthStep> {
	const data = await api.post("/auth/login", authStepDataSchema, payload);
	const step = parsePasswordAuthStep(data);
	if (step.kind === "session") {
		setAccessToken(step.session.access_token);
	}
	return step;
}

export async function register(payload: RegisterRequest): Promise<PasswordAuthStep> {
	const data = await api.post("/auth/register", authStepDataSchema, payload);
	const step = parsePasswordAuthStep(data);
	if (step.kind === "session") {
		setAccessToken(step.session.access_token);
	}
	return step;
}

export async function verifyLogin(payload: VerifyEmailCode) {
	const session = await api.post("/auth/login/verify", SessionSchema, payload);
	setAccessToken(session.access_token);
	return session;
}

export async function beginLoginWebAuthn(
	challenge_id: string,
): Promise<BeginLoginWebAuthnResponse> {
	return api.post("/auth/login/mfa/webauthn/begin", BeginLoginWebAuthnResponseSchema, {
		challenge_id,
	});
}

export async function verifyLoginMfa(payload: VerifyLoginMfaRequest) {
	const body = VerifyLoginMfaRequestSchema.parse(payload);
	const session = await api.post("/auth/login/mfa/verify", SessionSchema, body);
	setAccessToken(session.access_token);
	return session;
}

export async function verifyRegister(payload: VerifyEmailCode) {
	const session = await api.post("/auth/register/verify", SessionSchema, payload);
	setAccessToken(session.access_token);
	return session;
}

export async function resendLoginCode(payload: ResendEmailCode) {
	return api.post("/auth/login/resend", ResendEmailCodeResponseSchema, payload);
}

export async function resendRegisterCode(payload: ResendEmailCode) {
	return api.post("/auth/register/resend", ResendEmailCodeResponseSchema, payload);
}

export async function logout() {
	try {
		await api.post("/auth/logout", OkSchema, {});
	} finally {
		setAccessToken(null);
	}
}

export async function forgotPasswordRequest(payload: ForgotPasswordRequest) {
	return api.post("/auth/forgot-password", ForgotPasswordRequestResponseSchema, payload);
}

export async function forgotPasswordVerify(payload: ForgotPasswordVerify) {
	return api.post("/auth/forgot-password/verify", ForgotPasswordVerifyResponseSchema, payload);
}

export async function forgotPasswordReset(payload: ForgotPasswordReset) {
	return api.post("/auth/forgot-password/reset", OkSchema, payload);
}

export async function getSession(): Promise<User | null> {
	const token = await ensureAccessToken();
	if (!token) {
		return null;
	}
	try {
		return await api.get("/auth/me", UserSchema);
	} catch (error) {
		if (error instanceof ApiError && error.code === "auth.unauthorized") {
			setAccessToken(null);
			return null;
		}
		// Stale access token, gateway blip, or refresh cookie out of sync — treat as signed out.
		if (
			error instanceof ApiError &&
			(error.code === "error.unknown" ||
				error.code === "error.invalid_response" ||
				error.code === "error.network")
		) {
			setAccessToken(null);
			return null;
		}
		throw error;
	}
}
