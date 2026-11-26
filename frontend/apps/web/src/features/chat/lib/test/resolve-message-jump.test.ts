import assert from "node:assert/strict";
import { describe, it } from "node:test";
import type { Message } from "@wichat/contracts/chat/message";
import { findMessageSeqInCache } from "../message-cache-window";

describe("findMessageSeqInCache", () => {
	it("returns seq when message is in cache", () => {
		const msg: Message = {
			id: "aaaaaaaaaaaaaaaaaaaaaaaa",
			channel_seq: 42,
			author_id: "u1",
			author_display_name: "A",
			author_avatar_url: "",
			body: "hi",
			created_at: "2026-01-01T00:00:00.000Z",
			edited_at: "",
			is_pinned: false,
			mine: false,
			reactions: [],
			mentioned_user_ids: [],
		};
		const seq = findMessageSeqInCache(
			{ pages: [{ messages: [msg], has_more: false }], pageParams: [undefined] },
			msg.id,
		);
		assert.equal(seq, 42);
	});

	it("returns 0 when missing", () => {
		assert.equal(findMessageSeqInCache(undefined, "bbbbbbbbbbbbbbbbbbbbbbbb"), 0);
	});
});
