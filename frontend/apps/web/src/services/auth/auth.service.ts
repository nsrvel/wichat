import {
	OkSchema,
	SessionSchema,
	SetupStatusSchema,
	UserSchema,
	type LoginRequest,
	type SetupRequest,
	type User,
} from "@wichat/contracts";
import { ApiError } from "@wichat/api-client";
import { env } from "@/env";
import { api, ensureAccessToken, setAccessToken } from "@/integrations/api-client";

export function googleStartUrl() {
	return new URL("/api/v1/auth/google", env.NEXT_PUBLIC_API_URL).toString();
}

export async function getSetupStatus() {
	return api.get("/auth/setup-status", SetupStatusSchema);
}

export async function login(payload: LoginRequest) {
	const session = await api.post("/auth/login", SessionSchema, payload);
	setAccessToken(session.access_token);
	return session;
}

export async function setup(payload: SetupRequest) {
	const session = await api.post("/auth/setup", SessionSchema, payload);
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

export async function getSession(): Promise<User> {
	const token = await ensureAccessToken();
	if (!token) {
		throw new ApiError("auth.unauthorized");
	}
	return api.get("/auth/me", UserSchema);
}
