import type { JSONContent } from "@tiptap/react";
import { MENTION_USER_ID_SOURCE } from "@/features/chat/lib/mention-chip-styles";
import {
	MENTION_LINK_PREFIX,
	MENTION_ROLE_LINK_PREFIX,
} from "@/features/chat/lib/mention-markdown";

function applyMarks(text: string, marks: JSONContent["marks"]): string {
	if (!marks || marks.length === 0) {
		return text;
	}
	let out = text;
	for (const mark of marks) {
		switch (mark.type) {
			case "code":
				out = `\`${out}\``;
				break;
			case "bold":
				out = `**${out}**`;
				break;
			case "italic":
				out = `*${out}*`;
				break;
			case "strike":
				out = `~~${out}~~`;
				break;
			case "underline":
				out = `<u>${out}</u>`;
				break;
			case "highlight": {
				const color = typeof mark.attrs?.color === "string" ? mark.attrs.color : "";
				if (color) {
					out = `<mark data-hl="${color}">${out}</mark>`;
				}
				break;
			}
			case "textColor": {
				const color = typeof mark.attrs?.color === "string" ? mark.attrs.color : "";
				if (color) {
					out = `<span data-color="${color}">${out}</span>`;
				}
				break;
			}
			case "fontSize": {
				const size =
					mark.attrs?.size === "large" || mark.attrs?.size === "small" ? mark.attrs.size : "";
				if (size) {
					out = `<span data-size="${size}">${out}</span>`;
				}
				break;
			}
			case "link": {
				const href = typeof mark.attrs?.href === "string" ? mark.attrs.href : "";
				if (href) {
					out = `[${out}](${href})`;
				}
				break;
			}
			default:
				break;
		}
	}
	return out;
}

function serializeInline(nodes: JSONContent[] | undefined): string {
	if (!nodes || nodes.length === 0) {
		return "";
	}
	return nodes
		.map((node) => {
			if (node.type === "text") {
				return applyMarks(node.text ?? "", node.marks);
			}
			if (node.type === "hardBreak") {
				return "\n";
			}
			if (node.type === "userMention") {
				const id = typeof node.attrs?.id === "string" ? node.attrs.id : "";
				const rawLabel = typeof node.attrs?.label === "string" ? node.attrs.label : "";
				const cleanLabel = rawLabel.replace(/^@+/, "");
				if (!id) {
					return `@${cleanLabel}`;
				}
				return `[@${cleanLabel}](${MENTION_LINK_PREFIX}${id})`;
			}
			if (node.type === "roleMention") {
				const id = typeof node.attrs?.id === "string" ? node.attrs.id : "";
				const rawLabel = typeof node.attrs?.label === "string" ? node.attrs.label : "";
				const cleanLabel = rawLabel.replace(/^@+/, "");
				if (!id) {
					return `@${cleanLabel}`;
				}
				return `[@${cleanLabel}](${MENTION_ROLE_LINK_PREFIX}${id})`;
			}
			return serializeInline(node.content);
		})
		.join("");
}

function serializeList(node: JSONContent, ordered: boolean, indent = ""): string {
	const items = node.content ?? [];
	return items
		.map((item, idx) => {
			const marker = ordered ? `${idx + 1}. ` : "- ";
			const children = item.content ?? [];
			const lines: string[] = [];
			children.forEach((child, childIdx) => {
				if (child.type === "bulletList") {
					lines.push(serializeList(child, false, `${indent}  `));
				} else if (child.type === "orderedList") {
					lines.push(serializeList(child, true, `${indent}  `));
				} else {
					const text = serializeInline(child.content);
					if (childIdx === 0) {
						lines.push(`${indent}${marker}${text}`);
					} else if (text) {
						lines.push(`${indent}  ${text}`);
					}
				}
			});
			return lines.join("\n");
		})
		.join("\n");
}

export type InlineReplyPayload = {
	id: string;
	author: string;
	time?: string;
	preview: string;
};

export type InlineAttachmentPayload = {
	id: string;
	url: string;
	filename: string;
	content_type: string;
	size_bytes: number;
};

const INLINE_REPLY_FENCE_RE = /```wichat-reply\n[\s\S]*?\n```/g;
const INLINE_ATTACHMENT_FENCE_RE = /```wichat-attachment\n[\s\S]*?\n```/g;

export function hasInlineReplyBlock(markdown: string): boolean {
	return markdown.includes("```wichat-reply\n");
}

const INLINE_STYLE_TAGS_RE =
	/<\/?(?:u|mark(?:\s+data-hl="[a-z]+")?|span(?:\s+data-(?:color="[a-z]+"|size="(?:large|small)"))?)>/g;

export function stripInlineReplyBlocks(markdown: string): string {
	return markdown
		.replace(INLINE_REPLY_FENCE_RE, "")
		.replace(INLINE_ATTACHMENT_FENCE_RE, "")
		.replace(INLINE_STYLE_TAGS_RE, "")
		.replace(/\n{3,}/g, "\n\n")
		.trim();
}

export function parseInlineAttachmentPayload(raw: string): InlineAttachmentPayload | null {
	try {
		const parsed = JSON.parse(raw.trim()) as Partial<InlineAttachmentPayload>;
		if (
			typeof parsed.id !== "string" ||
			typeof parsed.url !== "string" ||
			typeof parsed.filename !== "string" ||
			typeof parsed.content_type !== "string"
		) {
			return null;
		}
		return {
			id: parsed.id,
			url: parsed.url,
			filename: parsed.filename,
			content_type: parsed.content_type,
			size_bytes: typeof parsed.size_bytes === "number" ? parsed.size_bytes : 0,
		};
	} catch {
		return null;
	}
}

function serializeAttachmentBlock(node: JSONContent): string {
	const status = typeof node.attrs?.status === "string" ? node.attrs.status : "uploaded";
	const url = typeof node.attrs?.url === "string" ? node.attrs.url : "";
	if (status !== "uploaded" || !url) {
		return "";
	}
	const payload: InlineAttachmentPayload = {
		id: typeof node.attrs?.attachmentId === "string" ? node.attrs.attachmentId : "",
		url,
		filename: typeof node.attrs?.filename === "string" ? node.attrs.filename : "file",
		content_type:
			typeof node.attrs?.contentType === "string"
				? node.attrs.contentType
				: "application/octet-stream",
		size_bytes: typeof node.attrs?.sizeBytes === "number" ? node.attrs.sizeBytes : 0,
	};
	return `\`\`\`wichat-attachment\n${JSON.stringify(payload)}\n\`\`\``;
}

export function parseInlineReplyPayload(raw: string): InlineReplyPayload | null {
	try {
		const parsed = JSON.parse(raw.trim()) as Partial<InlineReplyPayload>;
		if (typeof parsed.author !== "string" || typeof parsed.preview !== "string") {
			return null;
		}
		return {
			id: typeof parsed.id === "string" ? parsed.id : "",
			author: parsed.author,
			time: typeof parsed.time === "string" ? parsed.time : undefined,
			preview: parsed.preview,
		};
	} catch {
		return null;
	}
}

function serializeReplyQuote(node: JSONContent): string {
	const payload: InlineReplyPayload = {
		id: typeof node.attrs?.messageId === "string" ? node.attrs.messageId : "",
		author: typeof node.attrs?.authorName === "string" ? node.attrs.authorName : "",
		time: typeof node.attrs?.timeLabel === "string" ? node.attrs.timeLabel : "",
		preview: typeof node.attrs?.bodyPreview === "string" ? node.attrs.bodyPreview : "",
	};
	return `\`\`\`wichat-reply\n${JSON.stringify(payload)}\n\`\`\``;
}

export function serializeTipTapToMarkdown(doc: JSONContent): {
	markdown: string;
} {
	const blocks = doc.content ?? [];
	const parts: string[] = [];

	for (const block of blocks) {
		if (block.type === "replyQuote") {
			parts.push(serializeReplyQuote(block));
			continue;
		}
		if (block.type === "attachmentBlock") {
			const serializedAttachment = serializeAttachmentBlock(block);
			if (serializedAttachment) {
				parts.push(serializedAttachment);
			}
			continue;
		}

		let serialized = "";
		switch (block.type) {
			case "paragraph":
				serialized = serializeInline(block.content);
				break;
			case "heading": {
				const level = typeof block.attrs?.level === "number" ? block.attrs.level : 2;
				serialized = `${"#".repeat(level)} ${serializeInline(block.content)}`;
				break;
			}
			case "bulletList":
				serialized = serializeList(block, false);
				break;
			case "orderedList":
				serialized = serializeList(block, true);
				break;
			case "blockquote": {
				const inner = (block.content ?? []).map((c) => serializeInline(c.content)).join("\n");
				serialized = inner
					.split("\n")
					.map((line) => `> ${line}`)
					.join("\n");
				break;
			}
			case "codeBlock": {
				const lang = typeof block.attrs?.language === "string" ? block.attrs.language : "";
				const code = serializeInline(block.content);
				serialized = `\`\`\`${lang}\n${code}\n\`\`\``;
				break;
			}
			case "horizontalRule":
				serialized = "---";
				break;
			default:
				serialized = serializeInline(block.content);
				break;
		}

		if (serialized.trim().length > 0) {
			parts.push(serialized);
		}
	}

	return {
		markdown: parts.join("\n\n").trim(),
	};
}

type TipTapMark = NonNullable<JSONContent["marks"]>[number];

const INLINE_TOKEN_RE =
	/(<u>|<\/u>|<mark data-hl="[a-z]+">|<\/mark>|<span data-color="[a-z]+">|<span data-size="(?:large|small)">|<\/span>)/g;

const INLINE_MENTION_MD_RE = new RegExp(
	`\\[@([^\\]]+)\\]\\(wichat-mention:(${MENTION_USER_ID_SOURCE})\\)`,
	"gi",
);
const INLINE_ROLE_MENTION_MD_RE = new RegExp(
	`\\[@([^\\]]+)\\]\\(wichat-mention-role:([0-9a-f]{24})\\)`,
	"gi",
);

type MentionToken = {
	index: number;
	length: number;
	node: JSONContent;
};

function mentionTokensForLine(line: string): MentionToken[] {
	const tokens: MentionToken[] = [];
	INLINE_MENTION_MD_RE.lastIndex = 0;
	for (const match of line.matchAll(INLINE_MENTION_MD_RE)) {
		const index = match.index ?? 0;
		const rawLabel = match[1] ?? "";
		tokens.push({
			index,
			length: match[0].length,
			node: {
				type: "userMention",
				attrs: { id: match[2] ?? "", label: rawLabel.replace(/^@/, "") },
			},
		});
	}
	INLINE_ROLE_MENTION_MD_RE.lastIndex = 0;
	for (const match of line.matchAll(INLINE_ROLE_MENTION_MD_RE)) {
		const index = match.index ?? 0;
		const rawLabel = match[1] ?? "";
		tokens.push({
			index,
			length: match[0].length,
			node: {
				type: "roleMention",
				attrs: { id: match[2] ?? "", label: rawLabel.replace(/^@/, ""), color: "#99aab5" },
			},
		});
	}
	tokens.sort((a, b) => a.index - b.index);
	return tokens;
}

function parseInlineLineWithMentions(line: string): JSONContent[] {
	const tokens = mentionTokensForLine(line);
	if (tokens.length === 0) {
		return parseStyledLineToNodes(line);
	}
	const nodes: JSONContent[] = [];
	let lastIndex = 0;
	for (const token of tokens) {
		if (token.index > lastIndex) {
			nodes.push(...parseStyledLineToNodes(line.slice(lastIndex, token.index)));
		}
		nodes.push(token.node);
		lastIndex = token.index + token.length;
	}
	if (lastIndex < line.length) {
		nodes.push(...parseStyledLineToNodes(line.slice(lastIndex)));
	}
	return nodes;
}

function parseStyledLineToNodes(line: string): JSONContent[] {
	const nodes: JSONContent[] = [];
	const markStack: Array<{ tag: "u" | "mark" | "span"; mark: TipTapMark }> = [];
	let lastIndex = 0;

	for (const match of line.matchAll(INLINE_TOKEN_RE)) {
		const token = match[0];
		const idx = match.index ?? 0;
		if (idx > lastIndex) {
			const text = line.slice(lastIndex, idx);
			if (text.length > 0) {
				const marks = markStack.map((item) => item.mark);
				nodes.push(marks.length > 0 ? { type: "text", text, marks } : { type: "text", text });
			}
		}

		if (token === "<u>") {
			markStack.push({ tag: "u", mark: { type: "underline" } });
		} else if (token.startsWith("<mark data-hl=")) {
			const color = token.slice(15, -2);
			markStack.push({ tag: "mark", mark: { type: "highlight", attrs: { color } } });
		} else if (token.startsWith("<span data-color=")) {
			const color = token.slice(18, -2);
			markStack.push({ tag: "span", mark: { type: "textColor", attrs: { color } } });
		} else if (token.startsWith("<span data-size=")) {
			const size = token.slice(17, -2);
			markStack.push({ tag: "span", mark: { type: "fontSize", attrs: { size } } });
		} else if (token === "</u>" || token === "</mark>" || token === "</span>") {
			const targetTag = token === "</u>" ? "u" : token === "</mark>" ? "mark" : "span";
			for (let i = markStack.length - 1; i >= 0; i -= 1) {
				if (markStack[i]?.tag === targetTag) {
					markStack.splice(i, 1);
					break;
				}
			}
		}

		lastIndex = idx + token.length;
	}

	if (lastIndex < line.length) {
		const text = line.slice(lastIndex);
		if (text.length > 0) {
			const marks = markStack.map((item) => item.mark);
			nodes.push(marks.length > 0 ? { type: "text", text, marks } : { type: "text", text });
		}
	}

	return nodes;
}

function paragraphFromText(text: string): JSONContent {
	const lines = text.split("\n");
	const content: JSONContent[] = [];
	lines.forEach((line, idx) => {
		if (idx > 0) {
			content.push({ type: "hardBreak" });
		}
		if (line.length > 0) {
			content.push(...parseInlineLineWithMentions(line));
		}
	});
	return content.length > 0 ? { type: "paragraph", content } : { type: "paragraph" };
}

/** Converts stored message Markdown back into TipTap JSONContent for inline editing. */
export function markdownToTipTapContent(markdown: string): JSONContent {
	const trimmed = markdown.trim();
	if (!trimmed) {
		return { type: "doc", content: [{ type: "paragraph" }] };
	}

	const content: JSONContent[] = [];
	const fenceRe = /```([\w-]*)\n([\s\S]*?)\n```/g;
	let cursor = 0;
	let match: RegExpExecArray | null = fenceRe.exec(trimmed);

	const pushTextSegment = (segment: string) => {
		const cleaned = segment.trim();
		if (!cleaned) {
			return;
		}
		for (const para of cleaned.split(/\n{2,}/)) {
			const lines = para.split("\n");
			if (lines.every((line) => line.startsWith("> ") || line === ">")) {
				const quoteText = lines
					.map((line) => (line.startsWith("> ") ? line.slice(2) : ""))
					.join("\n");
				content.push({
					type: "blockquote",
					content: [paragraphFromText(quoteText)],
				});
			} else {
				content.push(paragraphFromText(para));
			}
		}
	};

	while (match) {
		if (match.index > cursor) {
			pushTextSegment(trimmed.slice(cursor, match.index));
		}
		const lang = match[1] ?? "";
		const body = match[2] ?? "";
		if (lang === "wichat-reply") {
			const reply = parseInlineReplyPayload(body);
			if (reply) {
				content.push({
					type: "replyQuote",
					attrs: {
						messageId: reply.id,
						authorName: reply.author,
						timeLabel: reply.time ?? "",
						bodyPreview: reply.preview,
					},
				});
			}
		} else if (lang === "wichat-attachment") {
			const att = parseInlineAttachmentPayload(body);
			if (att) {
				content.push({
					type: "attachmentBlock",
					attrs: {
						localId: att.id,
						attachmentId: att.id,
						url: att.url,
						previewUrl: "",
						filename: att.filename,
						contentType: att.content_type,
						sizeBytes: att.size_bytes,
						status: "uploaded",
						errorMessage: "",
					},
				});
			}
		} else {
			content.push({
				type: "codeBlock",
				attrs: { language: lang },
				content: body ? [{ type: "text", text: body }] : [],
			});
		}
		cursor = match.index + match[0].length;
		match = fenceRe.exec(trimmed);
	}

	if (cursor < trimmed.length) {
		pushTextSegment(trimmed.slice(cursor));
	}

	const lastType = content[content.length - 1]?.type;
	if (content.length === 0 || lastType === "replyQuote" || lastType === "attachmentBlock") {
		content.push({ type: "paragraph" });
	}

	return { type: "doc", content };
}
