import { type APIRequestContext, expect, test } from "@playwright/test";
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

test.describe("Category management, permissions & enforcement", () => {
	test("owner can edit category name, delete category, and changes persist across reload", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "catedit", "Category Edit Owner");
		const group = await createGroupWithInvite(request, owner, "E2E Category Edit Group");

		// Seed a category and a channel inside it via API
		const catRes = await request.post(
			`${E2E_API_BASE}/groups/${encodeURIComponent(group.id)}/categories`,
			{
				headers: { Authorization: `Bearer ${owner.accessToken}` },
				data: { name: "Operations", is_private: false },
			},
		);
		expect(catRes.status()).toBe(201);
		const catJson = (await catRes.json()) as { data: { category: { id: string } } };
		const categoryId = catJson.data.category.id;

		const chRes = await request.post(
			`${E2E_API_BASE}/groups/${encodeURIComponent(group.id)}/channels`,
			{
				headers: { Authorization: `Bearer ${owner.accessToken}` },
				data: { name: "ops-alerts", type: "text_only", category_id: categoryId, is_private: false },
			},
		);
		expect(chRes.status()).toBe(201);

		await loginInBrowser(context, owner);
		await page.goto(`/channel/${encodeURIComponent(group.id)}`);
		await expect(page.getByRole("button", { name: /^operations$/i })).toBeVisible();

		// Open Category Settings via right-click context menu
		await page.getByRole("button", { name: /^operations$/i }).click({ button: "right" });
		await page.getByRole("menuitem", { name: /edit category/i }).click();

		const settingsDialog = page.getByRole("dialog");
		await expect(settingsDialog).toBeVisible();

		const nameInput = settingsDialog.getByRole("textbox").first();
		await nameInput.fill("Platform Ops");

		const patchPromise = page.waitForResponse(
			(res) =>
				res.url().includes(`/categories/${encodeURIComponent(categoryId)}`) &&
				res.request().method() === "PATCH",
		);
		await settingsDialog.getByRole("button", { name: /save changes/i }).click();
		const patchRes = await patchPromise;
		expect(patchRes.status()).toBe(200);

		// Close settings overlay (Escape) and verify rename persists across reload
		await page.keyboard.press("Escape");
		await expect(settingsDialog).toBeHidden();
		await expect(page.getByRole("button", { name: /^platform ops$/i })).toBeVisible();

		await page.reload();
		await expect(page.getByRole("button", { name: /^platform ops$/i })).toBeVisible();

		// Delete Category via context menu -> Delete Category dialog
		await page.getByRole("button", { name: /^platform ops$/i }).click({ button: "right" });
		await page.getByRole("menuitem", { name: /delete category/i }).click();

		const deleteDialog = page.getByRole("dialog");
		await expect(deleteDialog).toBeVisible();

		const deletePromise = page.waitForResponse(
			(res) =>
				res.url().includes(`/categories/${encodeURIComponent(categoryId)}`) &&
				res.request().method() === "DELETE",
		);
		await deleteDialog.getByRole("button", { name: /delete category/i }).click();
		const deleteRes = await deletePromise;
		expect(deleteRes.status()).toBe(200);

		await expect(page.getByRole("button", { name: /^platform ops$/i })).toBeHidden();
		// Channel inside deleted category remains accessible as uncategorized
		await page.reload();
		await expect(page.getByRole("button", { name: /^platform ops$/i })).toBeHidden();
		await expect(page.getByRole("button", { name: /ops-alerts/i })).toBeVisible();
	});

	test("private category & member permission overrides control visibility for regular members", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "catprivown", "Priv Cat Owner");
		const member = await createTestUser(request, "catprivmem", "Priv Cat Member");
		const group = await createGroupWithInvite(request, owner, "E2E Private Category Group");
		await redeemInvite(request, member, group.inviteCode);

		// Owner creates a private category + channel inside it
		const catRes = await request.post(
			`${E2E_API_BASE}/groups/${encodeURIComponent(group.id)}/categories`,
			{
				headers: { Authorization: `Bearer ${owner.accessToken}` },
				data: { name: "Executive", is_private: true },
			},
		);
		expect(catRes.status()).toBe(201);
		const catJson = (await catRes.json()) as { data: { category: { id: string } } };
		const categoryId = catJson.data.category.id;

		const chRes = await request.post(
			`${E2E_API_BASE}/groups/${encodeURIComponent(group.id)}/channels`,
			{
				headers: { Authorization: `Bearer ${owner.accessToken}` },
				data: { name: "exec-brief", type: "text_only", category_id: categoryId, is_private: false },
			},
		);
		expect(chRes.status()).toBe(201);

		// Member API & UI: private category and its channel are hidden
		const memberPayloadBefore = await fetchGroupChannels(request, member, group.id);
		expect(memberPayloadBefore.categories.some((c) => c.id === categoryId)).toBe(false);
		expect(memberPayloadBefore.channels.some((c) => c.name === "exec-brief")).toBe(false);

		// Owner opens Category Settings -> Permissions and grants View Channels to the member
		await loginInBrowser(context, owner);
		await page.goto(`/channel/${encodeURIComponent(group.id)}`);
		await expect(page.getByRole("button", { name: /^executive$/i })).toBeVisible();

		await page.getByRole("button", { name: /^executive$/i }).click({ button: "right" });
		await page.getByRole("menuitem", { name: /edit category/i }).click();

		const settingsDialog = page.getByRole("dialog");
		await expect(settingsDialog).toBeVisible();
		await settingsDialog.getByRole("button", { name: /^permissions$/i }).click();

		// Open Roles/Members popover trigger and select the member
		await settingsDialog.getByRole("button", { name: /add role or member/i }).click();
		await page.getByRole("button", { name: /priv cat member/i }).click();

		// Grant View Channels (Allow is the first permission row's Allow button)
		await settingsDialog
			.getByRole("button", { name: /^allow$/i })
			.first()
			.click();

		const upsertPromise = page.waitForResponse(
			(res) =>
				res.url().includes(`/categories/${encodeURIComponent(categoryId)}/permissions`) &&
				res.request().method() === "PUT",
		);
		await settingsDialog.getByRole("button", { name: /save changes/i }).click();
		const upsertRes = await upsertPromise;
		expect(upsertRes.status()).toBe(200);

		// Verify member now sees Executive & exec-brief in API and UI
		const memberPayloadAfter = await fetchGroupChannels(request, member, group.id);
		expect(memberPayloadAfter.categories.some((c) => c.id === categoryId)).toBe(true);
		expect(memberPayloadAfter.channels.some((c) => c.name === "exec-brief")).toBe(true);

		// Switch browser session to member and verify sidebar visibility
		await loginInBrowser(context, member);
		await page.goto(`/channel/${encodeURIComponent(group.id)}`);
		await expect(page.getByRole("button", { name: /^executive$/i })).toBeVisible();
		await expect(page.getByRole("button", { name: /exec-brief/i })).toBeVisible();
	});
});
