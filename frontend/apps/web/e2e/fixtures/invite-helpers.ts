import { expect, type APIRequestContext } from "@playwright/test";
import { E2E_API_BASE, type TestUser } from "./api-helpers";
import { e2eTurnstileBody, e2eTurnstileQuery } from "./invite-turnstile";

type Envelope<T> = { data: T };

async function parseEnvelope<T>(res: Awaited<ReturnType<APIRequestContext["get"]>>): Promise<T> {
	const text = await res.text();
	expect(res.ok(), `API ${res.url()} failed (${res.status()}): ${text}`).toBeTruthy();
	return (JSON.parse(text) as Envelope<T>).data;
}

export async function fetchGroupInviteLink(
	request: APIRequestContext,
	owner: TestUser,
	groupId: string,
): Promise<{ token: string; url: string; expires_at: string }> {
	const turnstileQs = await e2eTurnstileQuery(request);
	const res = await request.get(
		`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/invite-link${turnstileQs ? `?${turnstileQs.slice(1)}` : ""}`,
		{ headers: { Authorization: `Bearer ${owner.accessToken}` } },
	);
	return parseEnvelope(res);
}

export async function previewInviteApi(request: APIRequestContext, token: string) {
	const turnstileQs = await e2eTurnstileQuery(request);
	const res = await request.get(
		`${E2E_API_BASE}/invites/preview?invite=${encodeURIComponent(token)}${turnstileQs}`,
	);
	const text = await res.text();
	expect(res.ok(), text).toBeTruthy();
	return (JSON.parse(text) as Envelope<{ valid: boolean; group_name?: string; reason?: string }>)
		.data;
}

export async function sendGroupInvitesByUserId(
	request: APIRequestContext,
	owner: TestUser,
	groupId: string,
	userIds: string[],
) {
	const turnstile = await e2eTurnstileBody(request);
	const res = await request.post(`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/invites`, {
		headers: { Authorization: `Bearer ${owner.accessToken}` },
		data: { user_ids: userIds, emails: [], ...turnstile },
	});
	return parseEnvelope<{ added: number; skipped: number; emailed: number }>(res);
}

export async function listMyGroups(request: APIRequestContext, user: TestUser) {
	const res = await request.get(`${E2E_API_BASE}/groups/mine`, {
		headers: { Authorization: `Bearer ${user.accessToken}` },
	});
	const text = await res.text();
	expect(res.ok(), text).toBeTruthy();
	return (JSON.parse(text) as Envelope<{ groups: { id: string; name: string }[] }>).data.groups;
}
