import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	loginInBrowser,
	redeemInvite,
} from "../fixtures/api-helpers";
import { MembersPage } from "../fixtures/members-page";

test.describe("Member ban moderation", () => {
	test("owner can temporary ban, permanent ban, and bulk ban members", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "owner", "Ban Owner");
		const tempTarget = await createTestUser(request, "tmp", "Temp Ban Target");
		const permTarget = await createTestUser(request, "prm", "Perm Ban Target");
		const bulkA = await createTestUser(request, "blka", "Bulk Target A");
		const bulkB = await createTestUser(request, "blkb", "Bulk Target B");

		const group = await createGroupWithInvite(request, owner, "E2E Ban Group");
		await redeemInvite(request, tempTarget, group.inviteCode);
		await redeemInvite(request, permTarget, group.inviteCode);
		await redeemInvite(request, bulkA, group.inviteCode);
		await redeemInvite(request, bulkB, group.inviteCode);

		await loginInBrowser(context, owner);
		const membersPage = new MembersPage(page);
		await membersPage.gotoMembers(group.id);

		// Temporary ban (1 hour)
		await membersPage.banMember(tempTarget.displayName, "1 hour", "Temp ban test");
		await expect(membersPage.rowByName(tempTarget.displayName)).toBeHidden();
		const tempRejoin = await redeemInvite(request, tempTarget, group.inviteCode);
		expect(tempRejoin.ok()).toBeFalsy();

		// Permanent ban
		await membersPage.banMember(permTarget.displayName, "Permanent", "Permanent ban test");
		await expect(membersPage.rowByName(permTarget.displayName)).toBeHidden();
		const permRejoin = await redeemInvite(request, permTarget, group.inviteCode);
		expect(permRejoin.ok()).toBeFalsy();

		// Bulk ban
		await membersPage.selectRow(bulkA.displayName);
		await membersPage.selectRow(bulkB.displayName);
		await membersPage.bulkBanSelected("Permanent", "Bulk ban test");
		await expect(membersPage.rowByName(bulkA.displayName)).toBeHidden();
		await expect(membersPage.rowByName(bulkB.displayName)).toBeHidden();
	});
});
