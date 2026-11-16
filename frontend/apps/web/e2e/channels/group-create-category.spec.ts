import { type APIRequestContext, expect, type Page, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	E2E_API_BASE,
	loginInBrowser,
	redeemInvite,
	type TestUser,
} from "../fixtures/api-helpers";

type ChannelsPayload = {
	channels: Array<{ id: string; name: string; type: string; category_id?: string | null }>;
	categories: Array<{ id: string; name: string; is_private: boolean; position: number }>;
};

async function fetchGroupChannels(
	request: APIRequestContext,
	user: TestUser,
	groupId: string,
): Promise<ChannelsPayload> {
	const res = await request.get(`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/channels`, {
		headers: { Authorization: `Bearer ${user.accessToken}` },
	});
	expect(res.ok(), `GET /groups/${groupId}/channels failed (${res.status()})`).toBeTruthy();
	const json = (await res.json()) as { data: ChannelsPayload };
	return json.data;
}

async function openGroupHeaderMenu(page: Page, groupName: string): Promise<void> {
	await page.getByRole("button", { name: new RegExp(groupName, "i") }).click();
}

test.describe("Group create category & starter sidebar", () => {
	test("new group shows permanent categories (Quick views, Favorites, Chats), starter Text Channels & Voice Channels, and 2 General channels", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "catstart", "Starter Owner");
		const groupName = "E2E Starter Categories Group";
		const group = await createGroupWithInvite(request, owner, groupName);

		const payload = await fetchGroupChannels(request, owner, group.id);
		expect(payload.categories.map((c) => c.name)).toEqual(["Text Channels", "Voice Channels"]);
		const generals = payload.channels.filter((c) => c.name === "General");
		expect(generals).toHaveLength(2);
		expect(generals.map((c) => c.type).sort()).toEqual(["text_only", "voice_only"]);

		await loginInBrowser(context, owner);
		await page.goto(`/channel/${encodeURIComponent(group.id)}`);

		// Permanent categories only: Quick views, Favorites, Chats
		await expect(page.getByRole("button", { name: /^quick views$/i })).toBeVisible();
		await expect(page.getByRole("button", { name: /^favorites$/i })).toBeVisible();
		await expect(page.getByRole("button", { name: /^chats$/i })).toBeVisible();

		// Legacy "Channels" category is removed
		await expect(page.getByRole("button", { name: /^channels$/i })).toBeHidden();

		// Starter categories & 2 General channels
		await expect(page.getByRole("button", { name: /^text channels$/i })).toBeVisible();
		await expect(page.getByRole("button", { name: /^voice channels$/i })).toBeVisible();
		await expect(page.getByRole("button", { name: /^general$/i })).toHaveCount(2);
	});

	test("owner can create public and private categories and they persist across reload", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "catcreate", "Create Cat Owner");
		const groupName = "E2E Create Category Group";
		const group = await createGroupWithInvite(request, owner, groupName);

		await loginInBrowser(context, owner);
		await page.goto(`/channel/${encodeURIComponent(group.id)}`);

		// Create public category "Engineering"
		await openGroupHeaderMenu(page, groupName);
		await page.getByRole("menuitem", { name: /^create category$/i }).click();

		const dialog = page.getByRole("dialog");
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole("heading", { name: /create category/i })).toBeVisible();

		const submitBtn = dialog.getByRole("button", { name: /^create category$/i });
		await expect(submitBtn).toBeDisabled();

		await dialog.getByRole("textbox").fill("Engineering");
		await expect(submitBtn).toBeEnabled();

		const createPublicResPromise = page.waitForResponse(
			(res) =>
				res.url().includes(`/api/v1/groups/${encodeURIComponent(group.id)}/categories`) &&
				res.request().method() === "POST",
		);
		await submitBtn.click();
		const publicRes = await createPublicResPromise;
		expect(publicRes.status()).toBe(201);
		await expect(dialog).toBeHidden();
		await expect(page.getByRole("button", { name: /^engineering$/i })).toBeVisible();

		// Create private category "Leadership"
		await openGroupHeaderMenu(page, groupName);
		await page.getByRole("menuitem", { name: /^create category$/i }).click();
		await expect(dialog).toBeVisible();
		await dialog.getByRole("textbox").fill("Leadership");
		await dialog.getByRole("switch").click();

		const createPrivateResPromise = page.waitForResponse(
			(res) =>
				res.url().includes(`/api/v1/groups/${encodeURIComponent(group.id)}/categories`) &&
				res.request().method() === "POST",
		);
		await dialog.getByRole("button", { name: /^create category$/i }).click();
		const privateRes = await createPrivateResPromise;
		expect(privateRes.status()).toBe(201);
		await expect(dialog).toBeHidden();
		await expect(page.getByRole("button", { name: /^leadership$/i })).toBeVisible();

		// Persists across reload
		await page.reload();
		await expect(page.getByRole("button", { name: /^engineering$/i })).toBeVisible();
		await expect(page.getByRole("button", { name: /^leadership$/i })).toBeVisible();

		const payload = await fetchGroupChannels(request, owner, group.id);
		const eng = payload.categories.find((c) => c.name === "Engineering");
		const lead = payload.categories.find((c) => c.name === "Leadership");
		expect(eng?.is_private).toBe(false);
		expect(lead?.is_private).toBe(true);
	});

	test("regular member without manage_channels cannot see Create Category menu item", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "catown2", "Cat Owner 2");
		const member = await createTestUser(request, "catmem2", "Cat Member 2");
		const groupName = "E2E Cat Perm Group";
		const group = await createGroupWithInvite(request, owner, groupName);
		await redeemInvite(request, member, group.inviteCode);

		await loginInBrowser(context, member);
		await page.goto(`/channel/${encodeURIComponent(group.id)}`);

		await openGroupHeaderMenu(page, groupName);
		await expect(page.getByRole("menuitem", { name: /^create category$/i })).toBeHidden();
	});
});
