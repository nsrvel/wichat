type MdastNode = {
	type: string;
	value?: string;
	children?: MdastNode[];
	data?: { hName?: string };
};

function transformChildren(children: MdastNode[]): MdastNode[] {
	const result: MdastNode[] = [];

	for (let i = 0; i < children.length; i++) {
		const node = children[i];
		if (!node) {
			continue;
		}

		if (node.type === "html" && node.value === "<u>") {
			const closeIndex = children.findIndex(
				(candidate, idx) => idx > i && candidate.type === "html" && candidate.value === "</u>",
			);
			if (closeIndex !== -1) {
				const inner = transformChildren(children.slice(i + 1, closeIndex));
				result.push({
					type: "underline",
					data: { hName: "u" },
					children: inner,
				});
				i = closeIndex;
				continue;
			}
		}

		if (node.children) {
			result.push({ ...node, children: transformChildren(node.children) });
		} else {
			result.push(node);
		}
	}

	return result;
}

/** Remark plugin that converts paired inline `<u>...</u>` HTML tokens into safe `<u>` elements. */
export function remarkUnderline() {
	return (tree: MdastNode) => {
		if (tree.children) {
			tree.children = transformChildren(tree.children);
		}
	};
}
