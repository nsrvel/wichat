import { createHash } from "node:crypto";
import { execSync } from "node:child_process";
import { expect, type APIRequestContext } from "@playwright/test";
import { E2E_API_BASE } from "./api-helpers";

type Envelope<T> = { data: T };

async function parseEnvelope<T>(res: Awaited<ReturnType<APIRequestContext["post"]>>): Promise<T> {
	const text = await res.text();
	const json = JSON.parse(text) as Envelope<T> & { error?: { code?: string } };
	return json.data;
}

async function parseErrorCode(
	res: Awaited<ReturnType<APIRequestContext["post"]>>,
): Promise<string> {
	const text = await res.text();
	const json = JSON.parse(text) as { error?: { code?: string } };
	return json.error?.code ?? "";
}

const DEV_E2E_OTP_KEY_PREFIX = "ms-user:dev:e2e-otp:";

function redisGet(key: string): string | null {
	const fromEnv = process.env.E2E_OTP_CODE?.trim();
	if (fromEnv) {
		return fromEnv;
	}
	const container = process.env.E2E_REDIS_CONTAINER?.trim() || "wichat-redis-1";
	const redisURL = process.env.E2E_REDIS_URL?.trim() || "redis://localhost:6379/0";
	try {
		const viaDocker = execSync(`docker exec ${container} redis-cli GET ${JSON.stringify(key)}`, {
			encoding: "utf8",
		}).trim();
		if (viaDocker && viaDocker !== "(nil)") {
			return viaDocker;
		}
	} catch {
		// fall through
	}
	try {
		const viaURL = execSync(`redis-cli -u ${JSON.stringify(redisURL)} GET ${JSON.stringify(key)}`, {
			encoding: "utf8",
		}).trim();
		if (viaURL && viaURL !== "(nil)") {
			return viaURL;
		}
	} catch {
		return null;
	}
	return null;
}

/** Reads the dev-only OTP shadow ms-user writes to Redis (development env only). */
export function readDevOtpForChallenge(challengeId: string): string | null {
	if (!challengeId.trim()) {
		return null;
	}
	return redisGet(`${DEV_E2E_OTP_KEY_PREFIX}${challengeId.trim()}`);
}

export function waitDevOtpForChallenge(challengeId: string, attempts = 20): string | null {
	for (let i = 0; i < attempts; i++) {
		const code = readDevOtpForChallenge(challengeId);
		if (code) {
			return code;
		}
		execSync("sleep 0.15");
	}
	return null;
}

export function e2eAuthSeedingAvailable(): boolean {
	if (process.env.E2E_OTP_CODE?.trim()) {
		return true;
	}
	try {
		const container = process.env.E2E_REDIS_CONTAINER?.trim() || "wichat-redis-1";
		execSync(`docker exec ${container} redis-cli PING`, { encoding: "utf8" });
		return true;
	} catch {
		return false;
	}
}

export async function loginPasswordChallenge(
	request: APIRequestContext,
	login: string,
	password: string,
): Promise<string> {
	const res = await request.post(`${E2E_API_BASE}/auth/login`, {
		data: { login, password, locale: "en" },
	});
	expect(res.ok(), await res.text()).toBeTruthy();
	const data = await parseEnvelope<{ challenge_id: string }>(res);
	expect(data.challenge_id).toBeTruthy();
	return data.challenge_id;
}

export async function verifyLoginCode(
	request: APIRequestContext,
	challengeId: string,
	code: string,
): Promise<{ access_token: string }> {
	const res = await request.post(`${E2E_API_BASE}/auth/login/verify`, {
		data: { challenge_id: challengeId, code },
	});
	expect(res.ok(), await res.text()).toBeTruthy();
	return parseEnvelope<{ access_token: string }>(res);
}

export async function expectLoginInvalidCredentials(
	request: APIRequestContext,
	login: string,
	password: string,
): Promise<void> {
	const res = await request.post(`${E2E_API_BASE}/auth/login`, {
		data: { login, password: password, locale: "en" },
	});
	expect(res.ok()).toBeFalsy();
	const code = await parseErrorCode(res);
	expect(code).toBe("auth.invalid_credentials");
}

export async function expectVerifyOtpRejected(
	request: APIRequestContext,
	challengeId: string,
	code: string,
): Promise<void> {
	const res = await request.post(`${E2E_API_BASE}/auth/login/verify`, {
		data: { challenge_id: challengeId, code },
	});
	expect(res.ok()).toBeFalsy();
}

export type RegisterPayload = {
	email: string;
	username: string;
	password: string;
	display_name: string;
	birth_date: string;
	locale: string;
};

export async function fetchPolicyInviteOnly(request: APIRequestContext): Promise<boolean> {
	const res = await request.get(`${E2E_API_BASE}/auth/setup-status`);
	expect(res.ok(), await res.text()).toBeTruthy();
	const data = await parseEnvelope<{ invite_only?: boolean }>(res);
	return data.invite_only === true;
}

export async function fetchTurnstileRequired(request: APIRequestContext): Promise<boolean> {
	const res = await request.get(`${E2E_API_BASE}/auth/setup-status`);
	expect(res.ok(), await res.text()).toBeTruthy();
	const data = await parseEnvelope<{ turnstile_required?: boolean }>(res);
	return data.turnstile_required === true;
}

export async function expectGoogleStartForbiddenForSignUp(
	request: APIRequestContext,
): Promise<void> {
	const res = await request.get(`${E2E_API_BASE}/auth/google?next=/sign-up`, { maxRedirects: 0 });
	expect([302, 303, 307]).toContain(res.status());
	const location = res.headers().location ?? "";
	expect(location).toContain("/forbidden");
}

export async function expectRegisterVerifyRequiresCaptcha(
	request: APIRequestContext,
	challengeId: string,
	code: string,
): Promise<void> {
	const res = await request.post(`${E2E_API_BASE}/auth/register/verify`, {
		data: { challenge_id: challengeId, code },
	});
	expect(res.ok()).toBeFalsy();
	const errCode = await parseErrorCode(res);
	expect(errCode).toBe("auth.captcha_failed");
}

export async function registerPasswordChallenge(
	request: APIRequestContext,
	body: RegisterPayload,
): Promise<string> {
	const res = await request.post(`${E2E_API_BASE}/auth/register`, { data: body });
	if (!res.ok() && (await isRegisterRateLimited(res))) {
		return "";
	}
	expect(res.ok(), await res.text()).toBeTruthy();
	const data = await parseEnvelope<{ challenge_id: string }>(res);
	expect(data.challenge_id).toBeTruthy();
	return data.challenge_id;
}

export async function expectRegisterRejected(
	request: APIRequestContext,
	body: RegisterPayload,
): Promise<void> {
	const res = await request.post(`${E2E_API_BASE}/auth/register`, { data: body });
	expect(res.ok()).toBeFalsy();
	const code = await parseErrorCode(res);
	expect(code).toBe("auth.registration_closed");
}

export async function expectRegisterUnavailable(
	request: APIRequestContext,
	body: RegisterPayload,
): Promise<void> {
	const res = await request.post(`${E2E_API_BASE}/auth/register`, { data: body });
	expect(res.ok()).toBeFalsy();
	const code = await parseErrorCode(res);
	expect(code).toBe("auth.register_unavailable");
}

export function isRegisterRateLimited(
	res: Awaited<ReturnType<APIRequestContext["post"]>>,
): Promise<boolean> {
	return parseErrorCode(res).then((code) => code === "error.rate_limited");
}

export async function verifyRegisterCode(
	request: APIRequestContext,
	challengeId: string,
	code: string,
): Promise<{ access_token: string; user: { id: string } }> {
	const res = await request.post(`${E2E_API_BASE}/auth/register/verify`, {
		data: { challenge_id: challengeId, code },
	});
	expect(res.ok(), await res.text()).toBeTruthy();
	return parseEnvelope<{ access_token: string; user: { id: string } }>(res);
}

export async function expectVerifyRegisterRejected(
	request: APIRequestContext,
	challengeId: string,
	code: string,
): Promise<void> {
	const res = await request.post(`${E2E_API_BASE}/auth/register/verify`, {
		data: { challenge_id: challengeId, code },
	});
	expect(res.ok()).toBeFalsy();
}

/** Stable Redis key segment for password-reset dev OTP (matches ms-user pkg/auth-redis EmailKey). */
export function devOtpKeyForEmail(email: string): string {
	const normalized = email.trim().toLowerCase();
	return `${DEV_E2E_OTP_KEY_PREFIX}${createHash("sha256").update(normalized).digest("hex")}`;
}

export function readDevOtpForPasswordReset(email: string): string | null {
	return redisGet(devOtpKeyForEmail(email));
}

export function waitDevOtpForPasswordReset(email: string, attempts = 20): string | null {
	for (let i = 0; i < attempts; i++) {
		const code = readDevOtpForPasswordReset(email);
		if (code) {
			return code;
		}
		execSync("sleep 0.15");
	}
	return null;
}

export async function forgotPasswordRequest(
	request: APIRequestContext,
	email: string,
	locale = "en",
): Promise<{ resend_cooldown_seconds: number }> {
	const res = await request.post(`${E2E_API_BASE}/auth/forgot-password`, {
		data: { email, locale },
	});
	expect(res.ok(), await res.text()).toBeTruthy();
	return parseEnvelope<{ resend_cooldown_seconds: number }>(res);
}

export async function expectForgotPasswordVerifyRequiresCaptcha(
	request: APIRequestContext,
	email: string,
	code: string,
): Promise<void> {
	const res = await request.post(`${E2E_API_BASE}/auth/forgot-password/verify`, {
		data: { email, code },
	});
	expect(res.ok()).toBeFalsy();
	const errCode = await parseErrorCode(res);
	expect(errCode).toBe("auth.captcha_failed");
}

export async function forgotPasswordVerify(
	request: APIRequestContext,
	email: string,
	code: string,
): Promise<{ reset_token: string }> {
	const res = await request.post(`${E2E_API_BASE}/auth/forgot-password/verify`, {
		data: { email, code },
	});
	expect(res.ok(), await res.text()).toBeTruthy();
	return parseEnvelope<{ reset_token: string }>(res);
}

export async function forgotPasswordReset(
	request: APIRequestContext,
	resetToken: string,
	password: string,
): Promise<void> {
	const res = await request.post(`${E2E_API_BASE}/auth/forgot-password/reset`, {
		data: { reset_token: resetToken, password },
	});
	expect(res.ok(), await res.text()).toBeTruthy();
}

export async function expectForgotPasswordWrongCode(
	request: APIRequestContext,
	email: string,
	code: string,
): Promise<void> {
	const res = await request.post(`${E2E_API_BASE}/auth/forgot-password/verify`, {
		data: { email, code },
	});
	expect(res.ok()).toBeFalsy();
	const errCode = await parseErrorCode(res);
	expect(errCode).toBe("auth.reset_code_invalid");
}
