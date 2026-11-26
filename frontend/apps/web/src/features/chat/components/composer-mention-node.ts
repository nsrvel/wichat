"use client";

import { mentionChipClassName } from "@/features/chat/lib/mention-chip-styles";
import { mergeAttributes, Node } from "@tiptap/react";

export const UserMentionNode = Node.create({
	name: "userMention",
	group: "inline",
	inline: true,
	atom: true,
	selectable: true,

	addAttributes() {
		return {
			id: { default: "" },
			label: { default: "" },
		};
	},

	parseHTML() {
		return [{ tag: "span[data-user-mention]" }];
	},

	renderHTML({ node, HTMLAttributes }) {
		const label = typeof node.attrs.label === "string" ? node.attrs.label : "";
		const id = typeof node.attrs.id === "string" ? node.attrs.id : "";
		return [
			"span",
			mergeAttributes(HTMLAttributes, {
				"data-user-mention": id,
				class: mentionChipClassName(id),
			}),
			label,
		];
	},

	renderText({ node }) {
		const label = typeof node.attrs.label === "string" ? node.attrs.label : "";
		return label;
	},
});