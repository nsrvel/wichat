import assert from "node:assert/strict";
import { describe, it } from "node:test";
import {
	mergeHttpMyPresence,
	resetPresenceMergeStateForTests,
	setPendingManualPresence,
	touchWsSelfPresence,
} from "../live-cache";

const degradedHttp = {
	manual_status: "online" as const,
	effective_status: "offline" as const,
	is_online: false,
	expires_at: "",
};

describe("mergeHttpMyPresence", () => {
	it("keeps fresher WS online state over degraded HTTP", () => {
		resetPresenceMergeStateForTests();
		const now = 1_700_000_000_000;
		touchWsSelfPresence({
			user_id: "u1",
			presence_status: "online",
			is_online: true,
		});
		const merged = mergeHttpMyPresence(degradedHttp, undefined, now + 5_000);
		assert.equal(merged.is_online, true);
		assert.equal(merged.effective_status, "online");
		assert.equal(merged.manual_status, "online");
	});

	it("respects pending manual status over stale HTTP manual", () => {
		resetPresenceMergeStateForTests();
		setPendingManualPresence("dnd");
		const merged = mergeHttpMyPresence(
			{
				manual_status: "online",
				effective_status: "online",
				is_online: true,
				expires_at: "",
			},
			undefined,
		);
		assert.equal(merged.manual_status, "dnd");
	});

	it("does not apply stale WS stamp past guard window", () => {
		resetPresenceMergeStateForTests();
		const now = 1_700_000_000_000;
		touchWsSelfPresence({
			user_id: "u1",
			presence_status: "online",
			is_online: true,
		});
		const merged = mergeHttpMyPresence(degradedHttp, undefined, now + 120_000);
		assert.equal(merged.is_online, false);
		assert.equal(merged.effective_status, "offline");
	});
});
