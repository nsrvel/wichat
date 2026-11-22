import { expect, test, type APIRequestContext } from "@playwright/test";
import { createGroupWithInvite, createTestUser, E2E_API_BASE, type TestUser } from "../fixtures/api-helpers";

type MessageRow = {
	id: string;
	body: string;
	channel_seq?: number;
};

async function listChannels(request: APIRequestContext, user: TestUser, groupId: string) {
	const res = await request.get(`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/channels`, {
		headers: { Authorization: `Bearer ${user.accessToken}` },
	});
	expect(res.ok(), `list channels failed: ${await res.text()}`).toBeTruthy();
	const json = (await res.json()) as {
		data: { channels: Array<{ id: string; type: string; name: string }> };
	};
	const text = json.data.channels.find((c) => c.type === "text_only" || c.type === "text");
	expect(text, "group should have a text channel").toBeTruthy();
	return text!;
}

async function postMessage(
	request: APIRequestContext,
	user: TestUser,
	groupId: string,
	channelId: string,
	body: string,
) {
	const res = await request.post(
		`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/messages`,
		{
			headers: { Authorization: `Bearer ${user.accessToken}` },
			data: { body },
		},
	);
	expect(res.ok(), `post message failed: ${await res.text()}`).toBeTruthy();
	const json = (await res.json()) as { data: { message: MessageRow } };
	return json.data.message;
}

async function listMessages(
	request: APIRequestContext,
	user: TestUser,
	groupId: string,
	channelId: string,
	query: Record<string, string>,
) {
	const qs = new URLSearchParams(query).toString();
	const res = await request.get(
		`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/messages?${qs}`,
		{ headers: { Authorization: `Bearer ${user.accessToken}` } },
	);
	expect(res.ok(), `list messages failed: ${await res.text()}`).toBeTruthy();
	const json = (await res.json()) as { data: { messages: MessageRow[]; has_more: boolean } };
	return json.data;
}

function seqValue(seq: number | undefined): number {
	return typeof seq === "number" && Number.isFinite(seq) && seq > 0 ? seq : 0;
}

test.describe("Reconnect gap replay (after_channel_seq)", () => {
	test("lists messages newer than cached high channel_seq", async ({ request }) => {
		const owner = await createTestUser(request, "gapreplay", "Gap Replay Owner");
		const group = await createGroupWithInvite(request, owner, "E2E Gap Replay Group");
		const channel = await listChannels(request, owner, group.id);

		const anchor = await postMessage(request, owner, group.id, channel.id, "gap anchor");
		const anchorSeq = seqValue(anchor.channel_seq);
		expect(anchorSeq).toBeGreaterThan(0);

		const bodies = ["gap one", "gap two", "gap three"];
		for (const body of bodies) {
			await postMessage(request, owner, group.id, channel.id, body);
		}

		const gapPage = await listMessages(request, owner, group.id, channel.id, {
			limit: "50",
			after_channel_seq: String(anchorSeq),
		});

		expect(gapPage.messages.length).toBe(3);
		const returnedBodies = gapPage.messages.map((m) => m.body).sort();
		expect(returnedBodies).toEqual([...bodies].sort());

		let prev = anchorSeq;
		for (const msg of gapPage.messages) {
			const seq = seqValue(msg.channel_seq);
			expect(seq).toBeGreaterThan(prev);
			prev = seq;
		}
	});
});