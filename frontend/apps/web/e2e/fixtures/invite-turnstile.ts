import type { APIRequestContext } from "@playwright/test";
import { fetchTurnstileRequired } from "./auth-helpers";

/** Cloudflare Turnstile always-pass token when gateway uses test secret keys. */
export const E2E_TURNSTILE_DUMMY = "XXXX.DUMMY.TOKEN.XXXX";

export async function e2eTurnstileQuery(request: APIRequestContext): Promise<string> {
	if (!(await fetchTurnstileRequired(request))) {
		return "";
	}
	const token = process.env.E2E_TURNSTILE_TOKEN?.trim() || E2E_TURNSTILE_DUMMY;
	return token ? `&turnstile_token=${encodeURIComponent(token)}` : "";
}

export async function e2eTurnstileBody(
	request: APIRequestContext,
): Promise<Record<string, string>> {
	if (!(await fetchTurnstileRequired(request))) {
		return {};
	}
	const token = process.env.E2E_TURNSTILE_TOKEN?.trim() || E2E_TURNSTILE_DUMMY;
	return token ? { turnstile_token: token } : {};
}
