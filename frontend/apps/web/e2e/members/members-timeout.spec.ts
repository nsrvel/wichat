import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	loginInBrowser,
	redeemInvite,
} from "../fixtures/api-helpers";
import { MembersPage } from "../fixtures/members-page";

test.describe("Member timeout moderation", () => {
	test("owner can timeout a member with portaled duration dropdown and remove timeout", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "owner", "Timeout Owner");
		const member = await createTestUser(request, "mem", "Timeout Target");

		const group = await createGroupWithInvite(request, owner, "E2E Timeout Group");
		await redeemInvite(request, member, group.inviteCode);

		await loginInBrowser(context, owner);
		const membersPage = new MembersPage(page);
		await membersPage.gotoMembers(group.id);

		await membersPage.timeoutMember(member.displayName, "5 mins", "Cooling down spam");

		// Timeout badge/label should be visible on row and persist across reload
		const row = membersPage.rowByName(member.displayName);
		await expect(row).toContainText(/timed out/i);

		await page.reload();
		await expect(membersPage.rowByName(member.displayName)).toContainText(/timed out/i);

		// Remove timeout
		await membersPage.removeTimeout(member.displayName);
		await expect(membersPage.rowByName(member.displayName)).not.toContainText(/timed out/i);
	});
});
