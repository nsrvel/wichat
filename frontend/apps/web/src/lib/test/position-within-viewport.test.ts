import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { positionWithinViewport } from "@/lib/position-within-viewport";

describe("positionWithinViewport", () => {
	it("flips below to above when not enough space under anchor", () => {
		const anchor = { top: 700, bottom: 740, left: 100, right: 300 };
		const pos = positionWithinViewport({
			anchor,
			width: 200,
			height: 120,
			gap: 6,
			padding: 8,
			viewportWidth: 800,
			viewportHeight: 800,
		});
		assert.equal(pos.side, "top");
		assert.equal(pos.top, 700 - 6 - 120);
	});

	it("shifts start-aligned panel left when it would overflow right", () => {
		const anchor = { top: 100, bottom: 132, left: 720, right: 780 };
		const pos = positionWithinViewport({
			anchor,
			width: 200,
			height: 80,
			align: "start",
			viewportWidth: 800,
			viewportHeight: 600,
			padding: 8,
		});
		assert.equal(pos.left, 592);
		assert.equal(pos.side, "bottom");
	});

	it("places end-aligned panel to the left of anchor right edge", () => {
		const anchor = { top: 100, bottom: 132, left: 40, right: 120 };
		const pos = positionWithinViewport({
			anchor,
			width: 168,
			height: 80,
			align: "end",
			viewportWidth: 400,
			viewportHeight: 600,
		});
		assert.equal(pos.left, 120 - 168);
	});
});