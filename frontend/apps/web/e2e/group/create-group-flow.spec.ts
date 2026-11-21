import { expect, test } from "@playwright/test";
import { createGroupApi, listGroupChannelCount } from "../fixtures/group-helpers";
import { createTestUser, loginInBrowser, type TestUser } from "../fixtures/api-helpers";

function uniqueName(prefix: string): string {
	return `E2E ${prefix} ${Date.now().toString(36)}`;
}

test.describe("create group flow", () => {
	test.describe.configure({ mode: "serial" });

	let owner: TestUser;
	let customGroupId = "";
	let customDefaultChannelId = "";

	test.beforeAll(async ({ request }) => {
		owner = await createTestUser(request, "grp_create", "Group Create Owner");
	});

	test("API creates custom workspace with default text channel", async ({ request }) => {
		const created = await createGroupApi(request, owner, uniqueName("Custom"), "custom");
		customGroupId = created.id;
		customDefaultChannelId = created.default_channel_id;
		const channels = await listGroupChannelCount(request, owner, created.id);
		expect(channels).toBeGreaterThanOrEqual(1);
	});

	test("API creates office template with expanded channel set", async ({ request }) => {
		const created = await createGroupApi(request, owner, uniqueName("Office"), "office");
		const channels = await listGroupChannelCount(request, owner, created.id);
		expect(channels).toBeGreaterThanOrEqual(6);
	});

	test("UI custom create navigates to default channel", async ({ page, context }) => {
		test.setTimeout(60_000);
		await loginInBrowser(context, owner);
		await page.goto(
			`/channel/${encodeURIComponent(customGroupId)}/${encodeURIComponent(customDefaultChannelId)}`,
		);
		await page
			.getByRole("button", { name: /create or join a group|buat atau gabung group/i })
			.click();
		await page.getByRole("button", { name: /create my own|buat sendiri/i }).click();

		const uiName = uniqueName("UI Custom");
		await page.locator('input[autocomplete="organization"]').fill(uiName);
		await page.getByRole("button", { name: /^create$|^buat$/i }).click();

		await expect(page).toHaveURL(/\/channel\/[0-9a-f]{24}\/[0-9a-f]{24}/, { timeout: 30_000 });
	});

	test("UI office template prefills name and lands on channel", async ({ page, context }) => {
		test.setTimeout(60_000);
		await loginInBrowser(context, owner);
		await page.goto(
			`/channel/${encodeURIComponent(customGroupId)}/${encodeURIComponent(customDefaultChannelId)}`,
		);
		await page
			.getByRole("button", { name: /create or join a group|buat atau gabung group/i })
			.click();
		await page.getByRole("button", { name: /^office$|^kantor$/i }).click();

		const nameInput = page.locator('input[autocomplete="organization"]');
		await expect(nameInput).not.toHaveValue("");
		await page.getByRole("button", { name: /^create$|^buat$/i }).click();

		await expect(page).toHaveURL(/\/channel\/[0-9a-f]{24}\/[0-9a-f]{24}/, { timeout: 30_000 });
	});
});
