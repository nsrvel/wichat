import assert from "node:assert/strict";
import { describe, it } from "node:test";
import type { Message } from "@wichat/contracts/chat/message";
import {
	applyBoundedMessageCache,
	channelHistoryHasMore,
	countCachedMessages,
	type ChannelMessagesInfiniteData,
} from "../message-cache-window";

function msg(id: string, seq: number): Message {
	return {
		id,
		channel_seq: seq,
		author_id: "u1",
		author_display_name: "Alice",
		author_avatar_url: "",
		body: id,
		created_at: "2026-01-01T00:00:00.000Z",
		edited_at: "",
		is_pinned: false,
		mine: false,
		reactions: [],
	};
}

function makeData(count: number, hasMore = true): ChannelMessagesInfiniteData {
	const messages: Message[] = [];
	for (let i = 1; i <= count; i++) {
		messages.push(msg(`m-${i}`, i));
	}
	return {
		pages: [{ messages, has_more: hasMore }],
		pageParams: [undefined],
		meta: { historyStartReached: !hasMore, oldestTrimmed: false, newestTrimmed: false },
	};
}

describe("applyBoundedMessageCache", () => {
	it("leaves small caches untouched", () => {
		const input = makeData(50);
		const out = applyBoundedMessageCache(input, "near-bottom", 3_000);
		assert.equal(countCachedMessages(out), 50);
	});

	it("trims oldest when near bottom", () => {
		const page0 = { messages: [msg("n-1", 5001), msg("n-2", 5002)], has_more: false };
		const page1 = { messages: Array.from({ length: 2_999 }, (_, i) => msg(`o-${i}`, i + 1)), has_more: true };
		const input: ChannelMessagesInfiniteData = {
			pages: [page0, page1],
			pageParams: [undefined, { beforeChannelSeq: 1 }],
			meta: { historyStartReached: false, oldestTrimmed: false, newestTrimmed: false },
		};
		const out = applyBoundedMessageCache(input, "near-bottom", 3_000);
		assert.ok(countCachedMessages(out) <= 3_000);
		assert.equal(out.meta?.oldestTrimmed, true);
	});
});

describe("channelHistoryHasMore", () => {
	it("is false when history start reached", () => {
		const input = makeData(10, false);
		assert.equal(channelHistoryHasMore(input), false);
	});

	it("is true when oldest pages were trimmed", () => {
		const input = makeData(10, false);
		input.meta = { historyStartReached: false, oldestTrimmed: true, newestTrimmed: false };
		assert.equal(channelHistoryHasMore(input), true);
	});
});