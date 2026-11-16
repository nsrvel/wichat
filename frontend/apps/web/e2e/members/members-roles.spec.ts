import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createRoleAndAssign,
	createTestUser,
	loginInBrowser,
	redeemInvite,
} from "../fixtures/api-helpers";
import { MembersPage } from "../fixtures/members-page";

test.describe("Members roles assignment", () => {
	test("owner can assign and unassign a role inline and it persists across reload", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "owner", "Role Owner");
		const member = await createTestUser(request, "mem", "Role Target");

		const group = await createGroupWithInvite(request, owner, "E2E Roles Group");
		await redeemInvite(request, member, group.inviteCode);
		await createRoleAndAssign(request, owner, group.id, "Moderator", owner.id, []);

		await loginInBrowser(context, owner);
		const membersPage = new MembersPage(page);
		await membersPage.gotoMembers(group.id);

		await membersPage.toggleRoleOnRow(member.displayName, "Moderator");
		await expect(membersPage.rowByName(member.displayName)).toContainText("Moderator");

		await page.reload();
		await expect(membersPage.rowByName(member.displayName)).toContainText("Moderator");

		await membersPage.toggleRoleOnRow(member.displayName, "Moderator");
		await expect(membersPage.rowByName(member.displayName)).not.toContainText("Moderator");
	});
});
