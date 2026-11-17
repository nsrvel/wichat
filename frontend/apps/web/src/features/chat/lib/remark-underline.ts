import {
	FONT_SIZE_STYLES,
	HIGHLIGHT_COLOR_MAP,
	TEXT_COLOR_MAP,
} from "@/features/chat/lib/composer-text-marks";

type MdastNode = {
	type: string;
	value?: string;
	children?: MdastNode[];
	data?: {
		hName?: string;
		hProperties?: Record<string, string>;
	};
};

const OPEN_HIGHLIGHT_RE = /^<mark data-hl="([a-z]+)">$/;
const OPEN_COLOR_RE = /^<span data-color="([a-z]+)">$/;
const OPEN_SIZE_RE = /^<span data-size="(large|small)">$/;

function findClosingHtmlIndex(
	children: MdastNode[],
	startIndex: number,
	tagName: "u" | "mark" | "span",
): number {
	const openPrefix = `<${tagName}`;
	const closeTag = `</${tagName}>`;
	let depth = 1;

	for (let i = startIndex + 1; i < children.length; i++) {
		const candidate = children[i];
		if (!candidate || candidate.type !== "html" || !candidate.value) {
			continue;
		}
		if (candidate.value === `<${tagName}>` || candidate.value.startsWith(`${openPrefix} `)) {
			depth += 1;
		} else if (candidate.value === closeTag) {
			depth -= 1;
			if (depth === 0) {
				return i;
			}
		}
	}
	return -1;
}

function transformChildren(children: MdastNode[]): MdastNode[] {
	const result: MdastNode[] = [];

	for (let i = 0; i < children.length; i++) {
		const node = children[i];
		if (!node) {
			continue;
		}

		if (node.type === "html" && node.value === "<u>") {
			const closeIndex = findClosingHtmlIndex(children, i, "u");
			if (closeIndex !== -1) {
				result.push({
					type: "underline",
					data: { hName: "u" },
					children: transformChildren(children.slice(i + 1, closeIndex)),
				});
				i = closeIndex;
				continue;
			}
		}

		if (node.type === "html" && node.value) {
			const hlMatch = OPEN_HIGHLIGHT_RE.exec(node.value);
			if (hlMatch?.[1] && HIGHLIGHT_COLOR_MAP[hlMatch[1]]) {
				const closeIndex = findClosingHtmlIndex(children, i, "mark");
				if (closeIndex !== -1) {
					result.push({
						type: "highlightMark",
						data: {
							hName: "mark",
							hProperties: {
								style: `background-color: ${HIGHLIGHT_COLOR_MAP[hlMatch[1]]}; border-radius: 3px; padding: 0 2px; color: inherit;`,
							},
						},
						children: transformChildren(children.slice(i + 1, closeIndex)),
					});
					i = closeIndex;
					continue;
				}
			}

			const colorMatch = OPEN_COLOR_RE.exec(node.value);
			if (colorMatch?.[1] && TEXT_COLOR_MAP[colorMatch[1]]) {
				const closeIndex = findClosingHtmlIndex(children, i, "span");
				if (closeIndex !== -1) {
					result.push({
						type: "textColorMark",
						data: {
							hName: "span",
							hProperties: {
								style: `color: ${TEXT_COLOR_MAP[colorMatch[1]]};`,
							},
						},
						children: transformChildren(children.slice(i + 1, closeIndex)),
					});
					i = closeIndex;
					continue;
				}
			}

			const sizeMatch = OPEN_SIZE_RE.exec(node.value);
			if (sizeMatch?.[1] && (sizeMatch[1] === "large" || sizeMatch[1] === "small")) {
				const closeIndex = findClosingHtmlIndex(children, i, "span");
				if (closeIndex !== -1) {
					const sizeStyle = FONT_SIZE_STYLES[sizeMatch[1]];
					result.push({
						type: "fontSizeMark",
						data: {
							hName: "span",
							hProperties: {
								style: `font-size: ${sizeStyle.fontSize}; line-height: ${sizeStyle.lineHeight};`,
							},
						},
						children: transformChildren(children.slice(i + 1, closeIndex)),
					});
					i = closeIndex;
					continue;
				}
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

/** Remark plugin that converts paired allowlisted inline formatting tags (`<u>`, `<mark data-hl>`, `<span data-color>`, `<span data-size>`) into safe elements. */
export function remarkUnderline() {
	return (tree: MdastNode) => {
		if (tree.children) {
			tree.children = transformChildren(tree.children);
		}
	};
}
