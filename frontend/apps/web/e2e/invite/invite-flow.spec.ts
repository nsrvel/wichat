import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	loginInBrowser,
	redeemInvite,
	type TestUser,
} from "../fixtures/api-helpers";
import { createGroupApi } from "../fixtures/group-helpers";
import {
	fetchGroupInviteLink,
	listMyGroups,
	previewInviteApi,
	sendGroupInvitesByUserId,
} from "../fixtures/invite-helpers";
import { fetchTurnstileRequired } from "../fixtures/auth-helpers";
import { E2E_API_BASE } from "../fixtures/api-helpers";

function uniqueName(prefix: string): string {
	return `E2E ${prefix} ${Date.now().toString(36)}`;
}

test.describe("invite flow", () => {
	test.describe.configure({ mode: "serial" });

	let owner: TestUser;
	let joiner: TestUser;
	let groupId = "";
	let inviteToken = "";
	let defaultChannelId = "";

	test.beforeAll(async ({ request }) => {
		owner = await createTestUser(request, "inv_owner", "Invite Owner");
		joiner = await createTestUser(request, "inv_join", "Invite Joiner");
		const created = await createGroupApi(request, owner, uniqueName("Invite Hub"));
		groupId = created.id;
		defaultChannelId = created.default_channel_id;
		const link = await fetchGroupInviteLink(request, owner, groupId);
		inviteToken = link.token;
	});

	test("API invite link preview is valid for members", async ({ request }) => {
		const preview = await previewInviteApi(request, inviteToken);
		expect(preview.valid).toBe(true);
		expect(preview.group_name).toContain("Invite Hub");
	});

	test("redeem requires captcha when turnstile enabled", async ({ request }) => {
		test.skip(!(await fetchTurnstileRequired(request)), "Turnstile not enabled on gateway");
		const joinerOnly = await createTestUser(request, "inv_cap", "Captcha Joiner");
		const res = await request.post(`${E2E_API_BASE}/invites/redeem`, {
			headers: { Authorization: `Bearer ${joinerOnly.accessToken}` },
			data: { token: inviteToken },
		});
		expect(res.ok()).toBeFalsy();
		const body = (await res.json()) as { error?: { code?: string } };
		expect(body.error?.code).toBe("auth.captcha_failed");
	});

	test("API preview soft-fails for unknown token", async ({ request }) => {
		const preview = await previewInviteApi(request, "totally-invalid-token-xyz");
		expect(preview.valid).toBe(false);
		expect(preview.reason).toBe("not_found");
	});

	test("API send-by-user-id adds existing user without share token", async ({ request }) => {
		const third = await createTestUser(request, "inv_direct", "Direct Add");
		const result = await sendGroupInvitesByUserId(request, owner, groupId, [third.id]);
		expect(result.added).toBe(1);
		const groups = await listMyGroups(request, third);
		expect(groups.some((g) => g.id === groupId)).toBe(true);
	});

	test("API share link redeem joins new member", async ({ request }) => {
		const res = await redeemInvite(request, joiner, inviteToken);
		expect(res.ok()).toBeTruthy();
		const groups = await listMyGroups(request, joiner);
		expect(groups.some((g) => g.id === groupId)).toBe(true);
	});

	test("API share link remains redeemable for another joiner", async ({ request }) => {
		const another = await createTestUser(request, "inv_reuse", "Reuse Joiner");
		const res = await redeemInvite(request, another, inviteToken);
		expect(res.ok()).toBeTruthy();
	});

	test("UI join dialog accepts pasted invite link", async ({ page, context, request }) => {
		test.setTimeout(60_000);
		await loginInBrowser(context, joiner);
		await page.goto(
			`/channel/${encodeURIComponent(groupId)}/${encodeURIComponent(defaultChannelId)}`,
		);
		await page
			.getByRole("button", { name: /create or join a group|buat atau gabung group/i })
			.click();
		await page.getByRole("dialog").getByRole("button", { name: /^Join a group$|^Gabung group$/i }).click();
		await expect(page.locator("#group-invite-code")).toBeVisible();
		await expect(page.getByText(/^invite code$/i)).toHaveCount(0);
		const seeded = await createGroupWithInvite(request, owner, uniqueName("Join UI"));
		await page.locator("#group-invite-code").fill(`http://localhost:8080/invite/${seeded.inviteCode}`);
		await page
			.getByRole("dialog")
			.getByRole("button", { name: /^Join group$|^Gabung group$/i })
			.click();
		await expect(page).toHaveURL(/\/channel\/[0-9a-f]{24}\/[0-9a-f]{24}/, { timeout: 30_000 });
	});

	test("UI invite landing accepts when signed in", async ({ page, context, request }) => {
		test.setTimeout(60_000);
		const landingUser = await createTestUser(request, "inv_land", "Landing User");
		const landingGroup = await createGroupWithInvite(request, owner, uniqueName("Landing"));
		await loginInBrowser(context, landingUser);
		await page.goto(`/invite/${encodeURIComponent(landingGroup.inviteCode)}`);
		await expect(
			page.getByRole("heading", { name: landingGroup.name, level: 2 }),
		).toBeVisible({
			timeout: 20_000,
		});
		await page.getByRole("button", { name: /accept invite|terima undangan/i }).click();
		await expect(page).toHaveURL(/\/channel\/[0-9a-f]{24}\/[0-9a-f]{24}/, { timeout: 30_000 });
	});
});