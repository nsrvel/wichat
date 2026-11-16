import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	loginInBrowser,
	redeemInvite,
} from "../fixtures/api-helpers";
import { MembersPage } from "../fixtures/members-page";

test.describe("Group settings audit logs", () => {
	test("owner sees recorded group activities, expands details, and filters by action", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "auditown", "Audit Owner");
		const member = await createTestUser(request, "auditmem", "Audit Target");

		const groupName = "E2E Audit Group";
		const group = await createGroupWithInvite(request, owner, groupName);
		await redeemInvite(request, member, group.inviteCode);

		await loginInBrowser(context, owner);
		const membersPage = new MembersPage(page);
		await membersPage.gotoMembers(group.id);
		await membersPage.kickMember(member.displayName, "Audit kick test");

		// Open Group Settings -> Audit Logs
		await page.getByRole("button", { name: new RegExp(groupName, "i") }).click();
		await page.getByRole("menuitem", { name: /group settings/i }).click();
		await page.getByRole("button", { name: /audit logs/i }).click();

		await expect(page.getByRole("heading", { name: /audit log/i })).toBeVisible();
		await expect(page.getByText(/created the group/i)).toBeVisible();
		await expect(page.getByText(/kicked/i)).toBeVisible();

		// Expand kick row to inspect reason detail
		const kickRow = page.locator('[data-testid^="audit-log-row-"]', { hasText: /kicked/i }).first();
		await kickRow.getByTestId("audit-log-expand-trigger").click();
		await expect(kickRow.getByText(/Audit kick test/)).toBeVisible();

		// Filter by Action -> Member Kicked
		await page.getByRole("button", { name: /all actions/i }).click();
		await page.getByRole("option", { name: /member kicked/i }).click();
		await expect(page.getByText(/kicked/i)).toBeVisible();
		await expect(page.getByText(/created the group/i)).toBeHidden();
	});

	test("regular member without view_audit_log cannot see Audit Logs nav item", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "audown2", "Audit Owner 2");
		const member = await createTestUser(request, "audmem2", "Regular Member");

		const groupName = "E2E Audit Perm Group";
		const group = await createGroupWithInvite(request, owner, groupName);
		await redeemInvite(request, member, group.inviteCode);

		await loginInBrowser(context, member);
		await page.goto(`/channel/${encodeURIComponent(group.id)}`);
		await page.getByRole("button", { name: new RegExp(groupName, "i") }).click();
		await page.getByRole("menuitem", { name: /group settings/i }).click();

		await expect(page.getByRole("button", { name: /group profile/i })).toBeVisible();
		await expect(page.getByRole("button", { name: /audit logs/i })).toBeHidden();
	});
});
