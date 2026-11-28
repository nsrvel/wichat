import { expect, test } from "@playwright/test";
import { createTestUser, loginInBrowser } from "../fixtures/api-helpers";
import { createGroupApi } from "../fixtures/group-helpers";

test.describe("Shell search", () => {
	test("opens composite search on a group channel route", async ({ page, context, request }) => {
		const owner = await createTestUser(request, "search", "Search Owner");
		const group = await createGroupApi(request, owner, "E2E Shell Search");

		await loginInBrowser(context, owner);
		await page.goto(`/channel/${group.id}/${group.default_channel_id}`);

		const trigger = page.getByRole("button", { name: /search/i }).first();
		await trigger.click();

		const searchField = page.getByRole("searchbox");
		await expect(searchField).toBeVisible();
		await searchField.fill("gen");

		await expect(page.getByText(/channels|messages|files/i).first()).toBeVisible();

		await page.keyboard.press("Escape");
		await expect(searchField).toBeHidden();
	});
});
