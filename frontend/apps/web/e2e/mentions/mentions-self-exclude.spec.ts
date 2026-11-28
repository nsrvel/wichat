import { expect, test } from "@playwright/test";
import {
	createGroupWithInvite,
	createTestUser,
	loginInBrowser,
	redeemInvite,
} from "../fixtures/api-helpers";
import {
	listGroupChannels,
	listGroupMentions,
	openGroupChannel,
	openMentionPicker,
	postChannelMessage,
	userMentionMarkdown,
} from "../fixtures/mention-helpers";

test.describe("Mentions self-exclude", () => {
	test("self-mention does not land in mentions inbox and picker hides self", async ({
		page,
		context,
		request,
	}) => {
		const owner = await createTestUser(request, "mself", "Mention Self Owner");
		const peer = await createTestUser(request, "mpeer", "Mention Peer");
		const group = await createGroupWithInvite(request, owner, "E2E Mention Self Group");
		await redeemInvite(request, peer, group.inviteCode);

		const channel = await listGroupChannels(request, owner, group.id);
		const selfBody = `ping ${userMentionMarkdown(owner.displayName, owner.id)}`;
		await postChannelMessage(request, owner, group.id, channel.id, selfBody);

		const inbox = await listGroupMentions(request, owner, group.id);
		expect(inbox.hits.some((h) => h.body_preview.includes(owner.displayName))).toBe(false);

		await loginInBrowser(context, owner);
		await openGroupChannel(page, group.id);
		await openMentionPicker(page);
		await expect(page.getByTestId(`mention-suggest-user-${owner.id}`)).toBeHidden();
		await expect(page.getByTestId(`mention-suggest-user-${peer.id}`)).toBeVisible();
	});
});
