import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createRoleAndAssign,
	createTestUser,
	redeemInvite,
} from "../fixtures/api-helpers";
import {
	listGroupChannels,
	listGroupMentions,
	postChannelMessage,
	roleMentionMarkdown,
} from "../fixtures/mention-helpers";

test.describe("Mentions role expand inbox", () => {
	test("role mention pings role members and excludes author", async ({ request }) => {
		const owner = await createTestUser(request, "mrexp", "Role Expand Owner");
		const inRole = await createTestUser(request, "mrin", "Role Expand In");
		const outRole = await createTestUser(request, "mrout", "Role Expand Out");
		const group = await createGroupWithInvite(request, owner, "E2E Role Expand Group");
		await redeemInvite(request, inRole, group.inviteCode);
		await redeemInvite(request, outRole, group.inviteCode);

		const role = await createRoleAndAssign(request, owner, group.id, "PingMod", inRole.id, []);
		await request.patch(
			`${process.env.E2E_API_URL ?? "http://localhost:8080/api/v1"}/groups/${encodeURIComponent(group.id)}/roles/${encodeURIComponent(role.id)}`,
			{
				headers: { Authorization: `Bearer ${owner.accessToken}` },
				data: {
					name: role.name,
					color: "#5865F2",
					hoist: false,
					mentionable: true,
					permissions: [],
				},
			},
		);

		const channel = await listGroupChannels(request, owner, group.id);
		const body = `hello ${roleMentionMarkdown(role.name, role.id)}`;
		await postChannelMessage(request, owner, group.id, channel.id, body);

		const ownerInbox = await listGroupMentions(request, owner, group.id);
		expect(ownerInbox.hits.some((h) => h.body_preview.includes(role.name))).toBe(false);

		const inInbox = await listGroupMentions(request, inRole, group.id);
		expect(inInbox.hits.some((h) => h.body_preview.includes(role.name))).toBe(true);

		const outInbox = await listGroupMentions(request, outRole, group.id);
		expect(outInbox.hits.some((h) => h.body_preview.includes(role.name))).toBe(false);
	});
});
