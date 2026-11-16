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
	channels: Array<{
		id: string;
		name: string;
		type: string;
		is_private: boolean;
		category_id?: string | null;
	}>;
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

test.describe("Group create channel", () => {
	test("owner can create channels via category + and group header menu with type, private toggle, navigation, and reload persistence", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "chcreate", "Channel Owner");
		const groupName = "E2E Create Channel Group";
		const group = await createGroupWithInvite(request, owner, groupName);

		await loginInBrowser(context, owner);
		await page.goto(`/channel/${encodeURIComponent(group.id)}`);

		// 1. Open Create Channel via the first category "+" button (Text Channels)
		const addChannelBtn = page.getByRole("button", { name: /^create channel$/i }).first();
		await expect(addChannelBtn).toBeVisible();
		await addChannelBtn.click();

		const dialog = page.getByRole("dialog");
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole("heading", { name: /create channel/i })).toBeVisible();

		const submitBtn = dialog.getByRole("button", { name: /^create channel$/i });
		await expect(submitBtn).toBeDisabled();

		// Select Forum type and fill channel name
		await dialog.getByRole("radio", { name: /forum/i }).click();
		await dialog.getByRole("textbox").fill("design-forum");
		await expect(submitBtn).toBeEnabled();

		const createForumResPromise = page.waitForResponse(
			(res) =>
				res.url().includes(`/api/v1/groups/${encodeURIComponent(group.id)}/channels`) &&
				res.request().method() === "POST",
		);
		await submitBtn.click();
		const forumRes = await createForumResPromise;
		expect(forumRes.status()).toBe(201);
		const forumJson = (await forumRes.json()) as { data: { channel: { id: string } } };

		await expect(dialog).toBeHidden();
		await expect(page).toHaveURL(
			new RegExp(`/channel/${encodeURIComponent(group.id)}/${forumJson.data.channel.id}$`),
		);
		await expect(page.getByRole("button", { name: /^design-forum$/i })).toBeVisible();

		// 2. Create private Announcement (broadcast) channel via Group Header menu
		await openGroupHeaderMenu(page, groupName);
		await page.getByRole("menuitem", { name: /^create channel$/i }).click();
		await expect(dialog).toBeVisible();

		await dialog.getByRole("radio", { name: /announcement/i }).click();
		await dialog.getByRole("textbox").fill("vip-updates");
		await dialog.getByRole("switch").click();

		const createBroadcastResPromise = page.waitForResponse(
			(res) =>
				res.url().includes(`/api/v1/groups/${encodeURIComponent(group.id)}/channels`) &&
				res.request().method() === "POST",
		);
		await dialog.getByRole("button", { name: /^create channel$/i }).click();
		const broadcastRes = await createBroadcastResPromise;
		expect(broadcastRes.status()).toBe(201);

		await expect(dialog).toBeHidden();
		await expect(page.getByRole("button", { name: /^vip-updates$/i })).toBeVisible();

		// 3. Reload and verify persistence + backend fields
		await page.reload();
		await expect(page.getByRole("button", { name: /^design-forum$/i })).toBeVisible();
		await expect(page.getByRole("button", { name: /^vip-updates$/i })).toBeVisible();

		const payload = await fetchGroupChannels(request, owner, group.id);
		const textCategory = payload.categories.find((c) => c.name === "Text Channels");
		const forumChannel = payload.channels.find((c) => c.name === "design-forum");
		const broadcastChannel = payload.channels.find((c) => c.name === "vip-updates");

		expect(forumChannel).toBeTruthy();
		expect(forumChannel?.type).toBe("forum");
		expect(forumChannel?.is_private).toBe(false);
		expect(forumChannel?.category_id).toBe(textCategory?.id);

		expect(broadcastChannel).toBeTruthy();
		expect(broadcastChannel?.type).toBe("broadcast");
		expect(broadcastChannel?.is_private).toBe(true);
	});

	test("regular member without manage_channels cannot see Create Channel triggers", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "chown2", "Ch Owner 2");
		const member = await createTestUser(request, "chmem2", "Ch Member 2");
		const groupName = "E2E Channel Perm Group";
		const group = await createGroupWithInvite(request, owner, groupName);
		await redeemInvite(request, member, group.inviteCode);

		await loginInBrowser(context, member);
		await page.goto(`/channel/${encodeURIComponent(group.id)}`);

		await expect(page.getByRole("button", { name: /^create channel$/i })).toBeHidden();
		await openGroupHeaderMenu(page, groupName);
		await expect(page.getByRole("menuitem", { name: /^create channel$/i })).toBeHidden();
	});
});
