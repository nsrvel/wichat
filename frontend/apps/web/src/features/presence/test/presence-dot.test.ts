import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { memberPresenceDot, presenceDotFromFields } from "../presence-dot";

describe("presenceDotFromFields", () => {
	it("returns offline when not online", () => {
		assert.equal(presenceDotFromFields(false, "online"), "offline");
	});

	it("maps effective statuses when online", () => {
		assert.equal(presenceDotFromFields(true, "idle"), "idle");
		assert.equal(presenceDotFromFields(true, "dnd"), "dnd");
	});

	it("treats effective offline while is_online as offline dot", () => {
		assert.equal(presenceDotFromFields(true, "offline"), "offline");
	});
});

describe("memberPresenceDot", () => {
	it("uses member fields", () => {
		assert.equal(
			memberPresenceDot({ is_online: true, presence_status: "online" }),
			"online",
		);
		assert.equal(
			memberPresenceDot({ is_online: false, presence_status: "online" }),
			"offline",
		);
	});
});