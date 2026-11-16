import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	loginInBrowser,
	redeemInvite,
} from "../fixtures/api-helpers";
import { MembersPage } from "../fixtures/members-page";

test.describe("Members permissions and hierarchy", () => {
	test("owner cannot be moderated and regular member has no moderation actions", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "owner", "Hier Owner");
		const memberA = await createTestUser(request, "mema", "Regular Member A");
		const memberB = await createTestUser(request, "memb", "Regular Member B");

		const group = await createGroupWithInvite(request, owner, "E2E Permissions Group");
		await redeemInvite(request, memberA, group.inviteCode);
		await redeemInvite(request, memberB, group.inviteCode);

		// 1. Owner cannot see Kick/Ban/Timeout for themselves
		await loginInBrowser(context, owner);
		const membersPage = new MembersPage(page);
		await membersPage.gotoMembers(group.id);
		await membersPage.openRowActions(owner.displayName);
		await expect(page.getByRole("menuitem", { name: /^kick /i })).toBeHidden();
		await expect(page.getByRole("menuitem", { name: /^ban /i })).toBeHidden();
		await expect(page.getByRole("menuitem", { name: /^timeout /i })).toBeHidden();
		await page.keyboard.press("Escape");

		// 2. Regular member A cannot see moderation actions for member B or owner
		await loginInBrowser(context, memberA);
		await membersPage.gotoMembers(group.id);
		await membersPage.openRowActions(memberB.displayName);
		await expect(page.getByRole("menuitem", { name: /^kick /i })).toBeHidden();
		await expect(page.getByRole("menuitem", { name: /^ban /i })).toBeHidden();
		await expect(page.getByRole("menuitem", { name: /^timeout /i })).toBeHidden();
	});
});
