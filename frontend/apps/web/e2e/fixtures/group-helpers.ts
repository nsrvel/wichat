import { expect, type APIRequestContext } from "@playwright/test";
import { E2E_API_BASE, type TestUser } from "./api-helpers";

type Envelope<T> = { data: T };

export type CreatedGroup = {
	id: string;
	name: string;
	default_channel_id: string;
	role: string;
	effective_permissions?: string[];
};

async function parseEnvelope<T>(res: Awaited<ReturnType<APIRequestContext["post"]>>): Promise<T> {
	const text = await res.text();
	const json = JSON.parse(text) as Envelope<T> & { error?: { code?: string } };
	return json.data;
}

export async function createGroupApi(
	request: APIRequestContext,
	owner: TestUser,
	name: string,
	template?: "custom" | "office" | "project" | "study" | "gaming",
): Promise<CreatedGroup> {
	const body: { name: string; template?: string } = { name };
	if (template && template !== "custom") {
		body.template = template;
	}
	const res = await request.post(`${E2E_API_BASE}/groups`, {
		headers: { Authorization: `Bearer ${owner.accessToken}` },
		data: body,
	});
	expect(res.ok(), await res.text()).toBeTruthy();
	const data = await parseEnvelope<CreatedGroup>(res);
	expect(data.id).toMatch(/^[0-9a-f]{24}$/);
	expect(data.default_channel_id).toMatch(/^[0-9a-f]{24}$/);
	expect(data.role).toBe("owner");
	return data;
}

export async function listGroupChannelCount(
	request: APIRequestContext,
	owner: TestUser,
	groupId: string,
): Promise<number> {
	const res = await request.get(`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/channels`, {
		headers: { Authorization: `Bearer ${owner.accessToken}` },
	});
	expect(res.ok(), await res.text()).toBeTruthy();
	const data = await parseEnvelope<{ channels: unknown[] }>(res);
	return data.channels.length;
}