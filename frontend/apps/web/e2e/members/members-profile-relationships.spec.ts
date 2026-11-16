import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	loginInBrowser,
	redeemInvite,
} from "../fixtures/api-helpers";
import { MembersPage } from "../fixtures/members-page";

test.describe("Member profile and relationships", () => {
	test("opens member profile dialog and toggles Ignore and Block", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "owner", "Profile Owner");
		const member = await createTestUser(request, "mem", "Profile Target");

		const group = await createGroupWithInvite(request, owner, "E2E Profile Group");
		await redeemInvite(request, member, group.inviteCode);

		await loginInBrowser(context, owner);
		const membersPage = new MembersPage(page);
		await membersPage.gotoMembers(group.id);

		const profileDialog = await membersPage.openProfile(member.displayName);
		await expect(profileDialog).toContainText(member.username);

		// Ignore -> Unignore
		await profileDialog.getByRole("button", { name: /^ignore$/i }).click();
		const confirmIgnore = page.getByRole("dialog").last();
		await confirmIgnore.getByRole("button", { name: /^ignore$/i }).click();
		await expect(profileDialog.getByRole("button", { name: /^unignore$/i })).toBeVisible();

		await profileDialog.getByRole("button", { name: /^unignore$/i }).click();
		await expect(profileDialog.getByRole("button", { name: /^ignore$/i })).toBeVisible();

		// Block -> Unblock
		await profileDialog.getByRole("button", { name: /^block$/i }).click();
		const confirmBlock = page.getByRole("dialog").last();
		await confirmBlock.getByRole("button", { name: /^block$/i }).click();
		await expect(profileDialog.getByRole("button", { name: /^unblock$/i })).toBeVisible();

		await profileDialog.getByRole("button", { name: /^unblock$/i }).click();
		await expect(profileDialog.getByRole("button", { name: /^block$/i })).toBeVisible();
	});
});
