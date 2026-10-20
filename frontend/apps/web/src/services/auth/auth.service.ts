import {
	AuthChallengeSchema,
	ForgotPasswordRequestResponseSchema,
	ForgotPasswordVerifyResponseSchema,
	InstancePolicySchema,
	OkSchema,
	parsePasswordAuthStep,
	ResendEmailCodeResponseSchema,
	SessionSchema,
	UserSchema,
	type ForgotPasswordRequest,
	type ForgotPasswordReset,
	type ForgotPasswordVerify,
	type GuestRegisterRequest,
	type LoginRequest,
	type PasswordAuthStep,
	type RegisterRequest,
	type ResendEmailCode,
	type User,
	type VerifyEmailCode,
} from "@wichat/contracts";
import { z } from "zod";
import { ApiError } from "@wichat/api-client/api-error";
import { env } from "@/env";
import { api, ensureAccessToken, setAccessToken } from "@/integrations/api-client";

export function googleStartUrl() {
	return new URL("/api/v1/auth/google", env.NEXT_PUBLIC_API_URL).toString();
}

export async function getInstancePolicy() {
	return api.get("/auth/setup-status", InstancePolicySchema);
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

export async function registerGuest(payload: GuestRegisterRequest) {
	const session = await api.post("/auth/register/guest", SessionSchema, payload);
	setAccessToken(session.access_token);
	return session;
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
	return api.post(
		"/auth/forgot-password/verify",
		ForgotPasswordVerifyResponseSchema,
		payload,
	);
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
		throw error;
	}
}
