import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { filterMentionSuggestMembers } from "@/features/chat/lib/mention-suggest-members";

describe("filterMentionSuggestMembers", () => {
	it("excludes self from suggestion list", () => {
		const members = [
			{ user_id: "self-id", display_name: "Self", username: "self" },
			{ user_id: "other-id", display_name: "Other", username: "other" },
		];
		const out = filterMentionSuggestMembers(members, "", "self-id");
		assert.equal(out.length, 1);
		assert.equal(out[0]?.id, "other-id");
	});
});
