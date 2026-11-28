import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createRoleAndAssign,
	createTestUser,
	loginInBrowser,
	redeemInvite,
} from "../fixtures/api-helpers";
import { openGroupChannel, openMentionPicker, patchGroupRole } from "../fixtures/mention-helpers";

test.describe("Mentions role picker permissions", () => {
	test("mentionable toggle and mention_everyone control role visibility", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "mrpick", "Role Picker Owner");
		const member = await createTestUser(request, "mrmem", "Role Picker Member");
		const group = await createGroupWithInvite(request, owner, "E2E Role Picker Group");
		await redeemInvite(request, member, group.inviteCode);

		const role = await createRoleAndAssign(request, owner, group.id, "E2E Mod", member.id, []);

		await loginInBrowser(context, member);
		await openGroupChannel(page, group.id);
		await openMentionPicker(page);
		await expect(page.getByTestId(`mention-suggest-role-${role.id}`)).toBeHidden();

		await patchGroupRole(request, owner, group.id, role.id, {
			name: role.name,
			color: "#5865F2",
			hoist: false,
			mentionable: true,
			permissions: [],
		});

		await page.reload();
		await openMentionPicker(page);
		await expect(page.getByTestId(`mention-suggest-role-${role.id}`)).toBeVisible();

		const elevated = await createRoleAndAssign(request, owner, group.id, "MentionAll", member.id, [
			"mention_everyone",
		]);
		await patchGroupRole(request, owner, group.id, role.id, {
			name: role.name,
			color: "#5865F2",
			hoist: false,
			mentionable: false,
			permissions: [],
		});

		await page.reload();
		await openMentionPicker(page);
		await expect(page.getByTestId(`mention-suggest-role-${role.id}`)).toBeVisible();
		await expect(page.getByTestId(`mention-suggest-role-${elevated.id}`)).toBeVisible();
	});
});
