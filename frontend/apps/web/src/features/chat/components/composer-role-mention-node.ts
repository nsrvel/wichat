"use client";

import { mergeAttributes, Node } from "@tiptap/react";

export const RoleMentionNode = Node.create({
	name: "roleMention",
	group: "inline",
	inline: true,
	atom: true,
	selectable: true,

	addAttributes() {
		return {
			id: { default: "" },
			label: { default: "" },
			color: { default: "#99aab5" },
		};
	},

	parseHTML() {
		return [{ tag: "span[data-role-mention]" }];
	},

	renderHTML({ node, HTMLAttributes }) {
		const rawLabel = typeof node.attrs.label === "string" ? node.attrs.label : "";
		const cleanLabel = rawLabel.replace(/^@+/, "");
		const color = typeof node.attrs.color === "string" ? node.attrs.color : "#99aab5";
		const id = typeof node.attrs.id === "string" ? node.attrs.id : "";
		return [
			"span",
			mergeAttributes(HTMLAttributes, {
				"data-role-mention": id,
				class: "mention-chip inline rounded-md px-1 font-medium no-underline",
				style: `background-color: color-mix(in oklab, ${color} 22%, transparent); color: ${color}`,
			}),
			`@${cleanLabel}`,
		];
	},

	renderText({ node }) {
		const rawLabel = typeof node.attrs.label === "string" ? node.attrs.label : "";
		return `@${rawLabel.replace(/^@+/, "")}`;
	},
});
