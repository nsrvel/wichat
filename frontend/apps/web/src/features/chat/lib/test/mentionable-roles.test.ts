import assert from "node:assert/strict";
import { describe, it } from "node:test";
import {
	isRoleVisibleInMentionPicker,
	mentionPickerRoles,
} from "@/features/chat/lib/mentionable-roles";

const baseRole = {
	id: "r1",
	slug: "mod",
	name: "Mod",
	color: "#5865F2",
	hoist: false,
	mentionable: false,
	is_default: false,
	is_system: false,
	position: 1,
	permissions: [],
	member_count: 0,
	created_at: "",
	updated_at: "",
};

describe("mentionable roles", () => {
	it("hides non-mentionable roles without mention_everyone", () => {
		assert.equal(isRoleVisibleInMentionPicker(baseRole, ["send_messages"]), false);
	});

	it("shows mentionable roles", () => {
		assert.equal(isRoleVisibleInMentionPicker({ ...baseRole, mentionable: true }, []), true);
	});

	it("mention_everyone bypasses mentionable flag", () => {
		assert.equal(isRoleVisibleInMentionPicker(baseRole, ["mention_everyone"]), true);
	});

	it("filters roles by query", () => {
		const roles = [
			{ ...baseRole, id: "a", name: "Alpha", mentionable: true },
			{ ...baseRole, id: "b", name: "Beta", mentionable: true },
		];
		const out = mentionPickerRoles(roles, [], "alp");
		assert.equal(out.length, 1);
		assert.equal(out[0]?.name, "Alpha");
	});
});
