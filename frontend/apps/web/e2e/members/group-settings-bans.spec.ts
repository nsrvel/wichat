import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	loginInBrowser,
	redeemInvite,
} from "../fixtures/api-helpers";
import { MembersPage } from "../fixtures/members-page";

test.describe("Group settings bans", () => {
	test("owner can view bans, search by username or reason, and unban a member", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "banowner", "Bans Owner");
		const tempUser = await createTestUser(request, "bantemp", "Temp Banned User");
		const permUser = await createTestUser(request, "banperm", "Perm Banned User");

		const groupName = "E2E Bans Settings Group";
		const group = await createGroupWithInvite(request, owner, groupName);
		await redeemInvite(request, tempUser, group.inviteCode);
		await redeemInvite(request, permUser, group.inviteCode);

		await loginInBrowser(context, owner);
		const membersPage = new MembersPage(page);
		await membersPage.gotoMembers(group.id);

		await membersPage.banMember(tempUser.displayName, "1 hour", "Spamming links");
		await membersPage.banMember(permUser.displayName, "Permanent", "Abusive behavior");

		// Open Group Settings -> Bans from the group header menu
		await page.getByRole("button", { name: new RegExp(groupName, "i") }).click();
		await page.getByRole("menuitem", { name: /group settings/i }).click();
		await page.getByRole("button", { name: /^bans$/i }).click();

		const bansTable = page.getByRole("table");
		await expect(bansTable).toBeVisible();
		await expect(bansTable.getByText(tempUser.displayName)).toBeVisible();
		await expect(bansTable.getByText("Spamming links")).toBeVisible();
		await expect(bansTable.getByText(permUser.displayName)).toBeVisible();
		await expect(bansTable.getByText("Abusive behavior")).toBeVisible();
		await expect(bansTable.getByText(/permanent/i)).toBeVisible();

		// Search by username
		const searchInput = page.getByPlaceholder(/search by username/i);
		await searchInput.fill(tempUser.username);
		await expect(bansTable.getByText(tempUser.displayName)).toBeVisible();
		await expect(bansTable.getByText(permUser.displayName)).toBeHidden();

		// Search by reason
		await searchInput.fill("Abusive");
		await expect(bansTable.getByText(permUser.displayName)).toBeVisible();
		await expect(bansTable.getByText(tempUser.displayName)).toBeHidden();

		// Clear search and unban tempUser
		await searchInput.fill("");
		await expect(bansTable.getByText(tempUser.displayName)).toBeVisible();

		const tempRow = page.locator("tbody tr", { hasText: tempUser.displayName }).first();
		await tempRow.getByRole("button", { name: /^unban$/i }).click();

		const confirmDialog = page.getByRole("dialog");
		await expect(confirmDialog).toBeVisible();
		await confirmDialog.getByRole("button", { name: /^unban$/i }).click();

		await expect(bansTable.getByText(tempUser.displayName)).toBeHidden();

		// Unbanned user can immediately rejoin with invite
		const rejoinRes = await redeemInvite(request, tempUser, group.inviteCode);
		expect(rejoinRes.ok()).toBeTruthy();
	});
});
