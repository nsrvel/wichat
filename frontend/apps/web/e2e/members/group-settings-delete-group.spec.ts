import { type APIRequestContext, expect, type Locator, type Page, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	E2E_API_BASE,
	loginInBrowser,
	redeemInvite,
	type TestUser,
} from "../fixtures/api-helpers";

async function openGroupSettings(page: Page, groupName: string): Promise<void> {
	await page.getByRole("button", { name: new RegExp(groupName, "i") }).click();
	await page.getByRole("menuitem", { name: /group settings/i }).click();
	await expect(page.getByRole("button", { name: /group profile/i })).toBeVisible();
}

async function openDeleteGroupConfirmDialog(page: Page, groupName: string): Promise<Locator> {
	await openGroupSettings(page, groupName);
	await page.getByRole("button", { name: /^delete group$/i }).click();
	const dialog = page.getByRole("dialog");
	await expect(dialog).toBeVisible();
	await expect(dialog.getByRole("heading", { name: /delete group/i })).toBeVisible();
	await expect(
		dialog.getByText(`Delete ${groupName} permanently? This cannot be undone.`),
	).toBeVisible();
	return dialog;
}

async function fetchMyGroupIds(request: APIRequestContext, user: TestUser): Promise<string[]> {
	const res = await request.get(`${E2E_API_BASE}/groups/mine`, {
		headers: { Authorization: `Bearer ${user.accessToken}` },
	});
	expect(res.ok(), `GET /groups/mine failed (${res.status()})`).toBeTruthy();
	const json = (await res.json()) as { data: { groups: Array<{ id: string }> } };
	return json.data.groups.map((g) => g.id);
}

test.describe("Group settings delete group", () => {
	test("owner can cancel group deletion from confirmation alert", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "delcancel", "Delete Cancel Owner");
		const groupName = "E2E Cancel Delete Group";
		const group = await createGroupWithInvite(request, owner, groupName);

		await loginInBrowser(context, owner);
		await page.goto(`/channel/${encodeURIComponent(group.id)}`);

		const dialog = await openDeleteGroupConfirmDialog(page, groupName);
		await dialog.getByRole("button", { name: /^cancel$/i }).click();
		await expect(dialog).toBeHidden();

		// Group Settings remains open and the group is still owned/present
		await expect(page.getByRole("button", { name: /^delete group$/i })).toBeVisible();
		const groupIds = await fetchMyGroupIds(request, owner);
		expect(groupIds).toContain(group.id);
	});

	test("owner can confirm delete group and is redirected to remaining group", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "delnext", "Delete Redirect Owner");
		const fallbackGroup = await createGroupWithInvite(request, owner, "E2E Fallback Group");
		const targetName = "E2E Target Delete Group";
		const targetGroup = await createGroupWithInvite(request, owner, targetName);

		await loginInBrowser(context, owner);
		await page.goto(`/channel/${encodeURIComponent(targetGroup.id)}`);

		const dialog = await openDeleteGroupConfirmDialog(page, targetName);
		const deleteResponsePromise = page.waitForResponse(
			(res) =>
				res.url().includes(`/api/v1/groups/${encodeURIComponent(targetGroup.id)}`) &&
				res.request().method() === "DELETE",
		);
		await dialog.getByRole("button", { name: /^delete$/i }).click();

		const deleteRes = await deleteResponsePromise;
		expect(deleteRes.ok()).toBeTruthy();

		await expect(dialog).toBeHidden();
		await expect(page).toHaveURL(new RegExp(`/channel/${fallbackGroup.id}/`), {
			timeout: 15_000,
		});

		const groupIds = await fetchMyGroupIds(request, owner);
		expect(groupIds).not.toContain(targetGroup.id);
		expect(groupIds).toContain(fallbackGroup.id);
	});

	test("owner deleting their only group is redirected to app home", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "delsolo", "Delete Solo Owner");
		const soloName = "E2E Solo Delete Group";
		const soloGroup = await createGroupWithInvite(request, owner, soloName);

		await loginInBrowser(context, owner);
		await page.goto(`/channel/${encodeURIComponent(soloGroup.id)}`);

		const dialog = await openDeleteGroupConfirmDialog(page, soloName);
		const deleteResponsePromise = page.waitForResponse(
			(res) =>
				res.url().includes(`/api/v1/groups/${encodeURIComponent(soloGroup.id)}`) &&
				res.request().method() === "DELETE",
		);
		await dialog.getByRole("button", { name: /^delete$/i }).click();

		const deleteRes = await deleteResponsePromise;
		expect(deleteRes.ok()).toBeTruthy();

		await expect(dialog).toBeHidden();
		await expect(page).toHaveURL(/\/home(?:\?|$)/, { timeout: 15_000 });

		const groupIds = await fetchMyGroupIds(request, owner);
		expect(groupIds).toEqual([]);
	});

	test("regular member cannot see Delete group action in group settings", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "delown", "Delete Perm Owner");
		const member = await createTestUser(request, "delmem", "Delete Perm Member");

		const groupName = "E2E Delete Perm Group";
		const group = await createGroupWithInvite(request, owner, groupName);
		await redeemInvite(request, member, group.inviteCode);

		await loginInBrowser(context, member);
		await page.goto(`/channel/${encodeURIComponent(group.id)}`);

		await openGroupSettings(page, groupName);
		await expect(page.getByRole("button", { name: /^delete group$/i })).toBeHidden();
	});
});
