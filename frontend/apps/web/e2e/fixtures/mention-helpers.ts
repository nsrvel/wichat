import { expect, type APIRequestContext, type Page } from "@playwright/test";
import { E2E_API_BASE, type TestUser } from "./api-helpers";

type Envelope<T> = { data: T };

async function parseEnvelope<T>(res: Awaited<ReturnType<APIRequestContext["get"]>>): Promise<T> {
	const text = await res.text();
	expect(res.ok(), `API ${res.url()} failed (${res.status()}): ${text}`).toBeTruthy();
	return (JSON.parse(text) as Envelope<T>).data;
}

export async function listGroupChannels(
	request: APIRequestContext,
	user: TestUser,
	groupId: string,
) {
	const res = await request.get(`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/channels`, {
		headers: { Authorization: `Bearer ${user.accessToken}` },
	});
	const data = await parseEnvelope<{ channels: Array<{ id: string; type: string; name: string }> }>(
		res,
	);
	const text = data.channels.find((c) => c.type === "text_only" || c.type === "text");
	expect(text, "group should have a text channel").toBeTruthy();
	return text!;
}

export async function postChannelMessage(
	request: APIRequestContext,
	user: TestUser,
	groupId: string,
	channelId: string,
	body: string,
) {
	const res = await request.post(
		`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/messages`,
		{
			headers: { Authorization: `Bearer ${user.accessToken}` },
			data: { body },
		},
	);
	expect(res.ok(), `post message failed: ${await res.text()}`).toBeTruthy();
}

export async function listGroupMentions(
	request: APIRequestContext,
	user: TestUser,
	groupId: string,
) {
	const res = await request.get(`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/mentions`, {
		headers: { Authorization: `Bearer ${user.accessToken}` },
	});
	return parseEnvelope<{ hits: Array<{ message_id: string; body_preview: string }> }>(res);
}

export async function patchGroupRole(
	request: APIRequestContext,
	owner: TestUser,
	groupId: string,
	roleId: string,
	body: Record<string, unknown>,
) {
	const res = await request.patch(
		`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/roles/${encodeURIComponent(roleId)}`,
		{
			headers: { Authorization: `Bearer ${owner.accessToken}` },
			data: body,
		},
	);
	expect(res.ok(), `patch role failed: ${await res.text()}`).toBeTruthy();
}

export async function listGroupRoles(request: APIRequestContext, user: TestUser, groupId: string) {
	const res = await request.get(`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/roles`, {
		headers: { Authorization: `Bearer ${user.accessToken}` },
	});
	return parseEnvelope<{ roles: Array<{ id: string; name: string; mentionable: boolean }> }>(res);
}

export async function openGroupChannel(page: Page, groupId: string) {
	await page.goto(`/channel/${encodeURIComponent(groupId)}`);
	await page.locator(".tiptap").first().waitFor({ state: "visible" });
}

export async function openMentionPicker(page: Page) {
	const editor = page.locator(".tiptap").first();
	await editor.click();
	await editor.pressSequentially("@", { delay: 40 });
	await page.getByRole("listbox").waitFor({ state: "visible" });
}

export function userMentionMarkdown(label: string, userId: string): string {
	return `[@${label}](wichat-mention:${userId})`;
}

export function roleMentionMarkdown(label: string, roleId: string): string {
	return `[@${label}](wichat-mention-role:${roleId})`;
}
