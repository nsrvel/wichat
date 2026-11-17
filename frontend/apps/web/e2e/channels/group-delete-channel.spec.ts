import { type APIRequestContext, expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	E2E_API_BASE,
	loginInBrowser,
	redeemInvite,
	type TestUser,
} from "../fixtures/api-helpers";

type ChannelRecord = {
	id: string;
	name: string;
	type: string;
	is_private: boolean;
	category_id?: string | null;
};

type ChannelsPayload = {
	channels: ChannelRecord[];
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

async function createChannelViaApi(
	request: APIRequestContext,
	user: TestUser,
	groupId: string,
	name: string,
): Promise<ChannelRecord> {
	const res = await request.post(`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/channels`, {
		headers: { Authorization: `Bearer ${user.accessToken}` },
		data: { name, type: "text" },
	});
	expect(res.ok(), `POST /groups/${groupId}/channels failed (${res.status()})`).toBeTruthy();
	const json = (await res.json()) as { data: { channel: ChannelRecord } };
	return json.data.channel;
}

test.describe("Group delete channel", () => {
	test("owner can delete channel via channel context menu (cancel, delete, redirect, and reload persistence)", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "chdelctx", "Delete Context Owner");
		const group = await createGroupWithInvite(request, owner, "E2E Delete Channel Context Group");
		const target = await createChannelViaApi(request, owner, group.id, "ctx-delete-ch");

		await loginInBrowser(context, owner);
		await page.goto(`/channel/${encodeURIComponent(group.id)}/${encodeURIComponent(target.id)}`);

		const targetBtn = page.getByRole("button", { name: /^ctx-delete-ch$/i });
		await expect(targetBtn).toBeVisible();

		// 1. Open context menu and test Cancel first
		await targetBtn.click({ button: "right" });
		await page.getByRole("menuitem", { name: /^delete channel$/i }).click();

		const confirmDialog = page.getByRole("dialog");
		await expect(confirmDialog).toBeVisible();
		await expect(confirmDialog.getByRole("heading", { name: /delete channel/i })).toBeVisible();
		await expect(confirmDialog).toContainText("ctx-delete-ch");

		await confirmDialog.getByRole("button", { name: /^cancel$/i }).click();
		await expect(confirmDialog).toBeHidden();
		await expect(targetBtn).toBeVisible();

		// 2. Re-open context menu and confirm deletion
		await targetBtn.click({ button: "right" });
		await page.getByRole("menuitem", { name: /^delete channel$/i }).click();
		await expect(confirmDialog).toBeVisible();

		const deleteResPromise = page.waitForResponse(
			(res) =>
				res.url().includes(`/api/v1/groups/${encodeURIComponent(group.id)}/channels/`) &&
				res.request().method() === "DELETE",
		);
		await confirmDialog.getByRole("button", { name: /^delete channel$/i }).click();
		const deleteRes = await deleteResPromise;
		expect(deleteRes.ok()).toBeTruthy();

		await expect(confirmDialog).toBeHidden();
		await expect(targetBtn).toHaveCount(0);
		await expect(page).not.toHaveURL(new RegExp(`/${encodeURIComponent(target.id)}$`));

		// 3. Reload and verify persistence in UI & backend
		await page.reload();
		await expect(page.getByRole("button", { name: /^ctx-delete-ch$/i })).toHaveCount(0);

		const payload = await fetchGroupChannels(request, owner, group.id);
		expect(payload.channels.some((c) => c.id === target.id)).toBe(false);
	});

	test("owner can delete channel via Edit Channel (Channel Settings) dialog", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "chdelset", "Delete Settings Owner");
		const group = await createGroupWithInvite(request, owner, "E2E Delete Channel Settings Group");
		const target = await createChannelViaApi(request, owner, group.id, "settings-delete-ch");

		await loginInBrowser(context, owner);
		await page.goto(`/channel/${encodeURIComponent(group.id)}/${encodeURIComponent(target.id)}`);

		const targetBtn = page.getByRole("button", { name: /^settings-delete-ch$/i });
		await expect(targetBtn).toBeVisible();

		// Open Channel Settings via Edit Channel context menu item
		await targetBtn.click({ button: "right" });
		await page.getByRole("menuitem", { name: /^edit channel$/i }).click();

		const settingsDialog = page.getByRole("dialog");
		await expect(settingsDialog).toBeVisible();

		// Click Delete Channel in Channel Settings sidebar
		await settingsDialog.getByRole("button", { name: /^delete channel$/i }).click();

		const confirmDialog = page.getByRole("dialog").filter({ hasText: "settings-delete-ch" });
		await expect(confirmDialog).toBeVisible();

		const deleteResPromise = page.waitForResponse(
			(res) =>
				res.url().includes(`/api/v1/groups/${encodeURIComponent(group.id)}/channels/`) &&
				res.request().method() === "DELETE",
		);
		await confirmDialog.getByRole("button", { name: /^delete channel$/i }).click();
		const deleteRes = await deleteResPromise;
		expect(deleteRes.ok()).toBeTruthy();

		await expect(targetBtn).toHaveCount(0);
		await expect(page).not.toHaveURL(new RegExp(`/${encodeURIComponent(target.id)}$`));

		await page.reload();
		await expect(page.getByRole("button", { name: /^settings-delete-ch$/i })).toHaveCount(0);

		const payload = await fetchGroupChannels(request, owner, group.id);
		expect(payload.channels.some((c) => c.id === target.id)).toBe(false);
	});

	test("regular member does not see Edit Channel or Delete Channel in channel context menu", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "chdelown", "Channel Owner");
		const member = await createTestUser(request, "chdelmem", "Regular Member");
		const group = await createGroupWithInvite(request, owner, "E2E Delete Channel Perm Group");
		const target = await createChannelViaApi(request, owner, group.id, "protected-ch");
		await redeemInvite(request, member, group.inviteCode);

		await loginInBrowser(context, member);
		await page.goto(`/channel/${encodeURIComponent(group.id)}/${encodeURIComponent(target.id)}`);

		const targetBtn = page.getByRole("button", { name: /^protected-ch$/i });
		await expect(targetBtn).toBeVisible();

		await targetBtn.click({ button: "right" });
		await expect(page.getByRole("menuitem", { name: /^edit channel$/i })).toHaveCount(0);
		await expect(page.getByRole("menuitem", { name: /^delete channel$/i })).toHaveCount(0);
	});
});
