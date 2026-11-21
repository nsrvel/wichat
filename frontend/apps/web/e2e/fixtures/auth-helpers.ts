import { execSync } from "node:child_process";
import { expect, type APIRequestContext } from "@playwright/test";
import { E2E_API_BASE } from "./api-helpers";

type Envelope<T> = { data: T };

async function parseEnvelope<T>(res: Awaited<ReturnType<APIRequestContext["post"]>>): Promise<T> {
	const text = await res.text();
	const json = JSON.parse(text) as Envelope<T> & { error?: { code?: string } };
	return json.data;
}

async function parseErrorCode(res: Awaited<ReturnType<APIRequestContext["post"]>>): Promise<string> {
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
		const viaDocker = execSync(
			`docker exec ${container} redis-cli GET ${JSON.stringify(key)}`,
			{ encoding: "utf8" },
		).trim();
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

export async function registerPasswordChallenge(
	request: APIRequestContext,
	body: RegisterPayload,
): Promise<string> {
	const res = await request.post(`${E2E_API_BASE}/auth/register`, { data: body });
	expect(res.ok(), await res.text()).toBeTruthy();
	const data = await parseEnvelope<{ challenge_id: string }>(res);
	expect(data.challenge_id).toBeTruthy();
	return data.challenge_id;
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