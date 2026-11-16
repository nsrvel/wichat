import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	loginInBrowser,
	redeemInvite,
} from "../fixtures/api-helpers";
import { MembersPage } from "../fixtures/members-page";

test.describe("Member kick moderation", () => {
	test("owner can kick a member with reason and kicked user can rejoin via invite", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "owner", "Kick Owner");
		const member = await createTestUser(request, "mem", "Kick Target");

		const group = await createGroupWithInvite(request, owner, "E2E Kick Group");
		await redeemInvite(request, member, group.inviteCode);

		await loginInBrowser(context, owner);
		const membersPage = new MembersPage(page);
		await membersPage.gotoMembers(group.id);

		await expect(membersPage.rowByName(member.displayName)).toBeVisible();
		await membersPage.kickMember(member.displayName, "Rule reminder");
		await expect(membersPage.rowByName(member.displayName)).toBeHidden();

		// Kicked member can rejoin using invite
		const rejoinRes = await redeemInvite(request, member, group.inviteCode);
		expect(rejoinRes.ok()).toBeTruthy();

		await page.reload();
		await expect(membersPage.rowByName(member.displayName)).toBeVisible();
	});
});
