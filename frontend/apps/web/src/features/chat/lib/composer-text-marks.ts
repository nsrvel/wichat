import { Mark, mergeAttributes } from "@tiptap/react";

export type ColorSwatch = {
	key: string;
	label: string;
	color: string;
};

export const HIGHLIGHT_COLORS: readonly ColorSwatch[] = [
	{ key: "rose", label: "Rose", color: "rgba(190, 82, 98, 0.45)" },
	{ key: "terracotta", label: "Orange", color: "rgba(186, 94, 58, 0.45)" },
	{ key: "amber", label: "Amber", color: "rgba(176, 126, 18, 0.45)" },
	{ key: "olive", label: "Olive", color: "rgba(110, 134, 18, 0.45)" },
	{ key: "green", label: "Green", color: "rgba(38, 128, 74, 0.45)" },
	{ key: "teal", label: "Teal", color: "rgba(34, 122, 122, 0.45)" },
	{ key: "slate", label: "Slate", color: "rgba(72, 92, 124, 0.5)" },
	{ key: "plum", label: "Plum", color: "rgba(104, 72, 108, 0.5)" },
] as const;

export const TEXT_COLORS: readonly ColorSwatch[] = [
	{ key: "red", label: "Red", color: "#d9535f" },
	{ key: "orange", label: "Orange", color: "#dc6439" },
	{ key: "yellow", label: "Yellow", color: "#f6c137" },
	{ key: "lime", label: "Lime", color: "#b6cf3b" },
	{ key: "green", label: "Green", color: "#36a662" },
	{ key: "teal", label: "Teal", color: "#3a8f8c" },
	{ key: "blue", label: "Blue", color: "#2d68cc" },
	{ key: "magenta", label: "Magenta", color: "#b44787" },
] as const;

export const FONT_SIZE_STYLES = {
	large: { fontSize: "17px", lineHeight: "1.45" },
	small: { fontSize: "12.5px", lineHeight: "1.45" },
} as const;

export type FontSizeKey = "large" | "medium" | "small";

export const HIGHLIGHT_COLOR_MAP: Record<string, string> = Object.fromEntries(
	HIGHLIGHT_COLORS.map((item) => [item.key, item.color]),
);

export const TEXT_COLOR_MAP: Record<string, string> = Object.fromEntries(
	TEXT_COLORS.map((item) => [item.key, item.color]),
);

declare module "@tiptap/react" {
	interface Commands<ReturnType> {
		highlightColor: {
			setHighlightColor: (colorKey: string) => ReturnType;
			unsetHighlightColor: () => ReturnType;
		};
		textColor: {
			setTextColor: (colorKey: string) => ReturnType;
			unsetTextColor: () => ReturnType;
		};
		fontSize: {
			setFontSize: (size: "large" | "small") => ReturnType;
			unsetFontSize: () => ReturnType;
		};
	}
}

export const HighlightMark = Mark.create({
	name: "highlight",

	addAttributes() {
		return {
			color: {
				default: null,
				parseHTML: (element) => element.getAttribute("data-hl"),
				renderHTML: (attributes) => {
					const key = typeof attributes.color === "string" ? attributes.color : "";
					const bg = HIGHLIGHT_COLOR_MAP[key];
					if (!bg) {
						return {};
					}
					return {
						"data-hl": key,
						style: `background-color: ${bg}; border-radius: 3px; padding: 0 2px; color: inherit;`,
					};
				},
			},
		};
	},

	parseHTML() {
		return [{ tag: "mark[data-hl]" }];
	},

	renderHTML({ HTMLAttributes }) {
		return ["mark", mergeAttributes(HTMLAttributes), 0];
	},

	addCommands() {
		return {
			setHighlightColor:
				(colorKey: string) =>
				({ commands }) =>
					HIGHLIGHT_COLOR_MAP[colorKey] ? commands.setMark(this.name, { color: colorKey }) : false,
			unsetHighlightColor:
				() =>
				({ commands }) =>
					commands.unsetMark(this.name),
		};
	},
});

export const TextColorMark = Mark.create({
	name: "textColor",

	addAttributes() {
		return {
			color: {
				default: null,
				parseHTML: (element) => element.getAttribute("data-color"),
				renderHTML: (attributes) => {
					const key = typeof attributes.color === "string" ? attributes.color : "";
					const fg = TEXT_COLOR_MAP[key];
					if (!fg) {
						return {};
					}
					return {
						"data-color": key,
						style: `color: ${fg};`,
					};
				},
			},
		};
	},

	parseHTML() {
		return [{ tag: "span[data-color]" }];
	},

	renderHTML({ HTMLAttributes }) {
		return ["span", mergeAttributes(HTMLAttributes), 0];
	},

	addCommands() {
		return {
			setTextColor:
				(colorKey: string) =>
				({ commands }) =>
					TEXT_COLOR_MAP[colorKey] ? commands.setMark(this.name, { color: colorKey }) : false,
			unsetTextColor:
				() =>
				({ commands }) =>
					commands.unsetMark(this.name),
		};
	},
});

export const FontSizeMark = Mark.create({
	name: "fontSize",

	addAttributes() {
		return {
			size: {
				default: null,
				parseHTML: (element) => element.getAttribute("data-size"),
				renderHTML: (attributes) => {
					const key =
						attributes.size === "large" ? "large" : attributes.size === "small" ? "small" : null;
					if (!key) {
						return {};
					}
					const style = FONT_SIZE_STYLES[key];
					return {
						"data-size": key,
						style: `font-size: ${style.fontSize}; line-height: ${style.lineHeight};`,
					};
				},
			},
		};
	},

	parseHTML() {
		return [{ tag: "span[data-size]" }];
	},

	renderHTML({ HTMLAttributes }) {
		return ["span", mergeAttributes(HTMLAttributes), 0];
	},

	addCommands() {
		return {
			setFontSize:
				(size: "large" | "small") =>
				({ commands }) =>
					commands.setMark(this.name, { size }),
			unsetFontSize:
				() =>
				({ commands }) =>
					commands.unsetMark(this.name),
		};
	},
});
