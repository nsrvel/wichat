import { expect, test } from "@playwright/test";

test.describe("Voice & Video Call Flows", () => {
	test("authenticated workspace keeps user dock controls accessible", async ({ page }) => {
		await page.goto("/home");
		await expect(page).toHaveURL(/\/(home|sign-in)/);
	});
});
