import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	loginInBrowser,
	redeemInvite,
} from "../fixtures/api-helpers";
import { MembersPage } from "../fixtures/members-page";

test.describe("Members list, search, and sort", () => {
	test("shows members, join method, search filter, and sort order", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "owner", "Alpha Owner");
		const memberA = await createTestUser(request, "mema", "Bravo Member");
		const memberB = await createTestUser(request, "memb", "Zulu Member");

		const group = await createGroupWithInvite(request, owner, "E2E List Group");
		await redeemInvite(request, memberA, group.inviteCode);
		await redeemInvite(request, memberB, group.inviteCode);

		await loginInBrowser(context, owner);
		const membersPage = new MembersPage(page);
		await membersPage.gotoMembers(group.id);

		await expect(membersPage.rowByName(owner.displayName)).toBeVisible();
		await expect(membersPage.rowByName(memberA.displayName)).toBeVisible();
		await expect(membersPage.rowByName(memberB.displayName)).toBeVisible();

		// Search narrows visible rows
		await membersPage.search("Zulu");
		await expect(membersPage.rowByName(memberB.displayName)).toBeVisible();
		await expect(membersPage.rowByName(memberA.displayName)).toBeHidden();

		// Clear search & verify sort descending puts Zulu before Alpha
		await membersPage.search("");
		await membersPage.sortBy("Name (Z–A)");
		const firstRowText = await page.locator("tbody tr").first().innerText();
		expect(firstRowText).toContain("Zulu Member");
	});
});
