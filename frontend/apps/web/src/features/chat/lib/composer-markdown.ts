import type { JSONContent } from "@tiptap/react";

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

export function stripInlineReplyBlocks(markdown: string): string {
	return markdown
		.replace(INLINE_REPLY_FENCE_RE, "")
		.replace(INLINE_ATTACHMENT_FENCE_RE, "")
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

function paragraphFromText(text: string): JSONContent {
	const lines = text.split("\n");
	const content: JSONContent[] = [];
	lines.forEach((line, idx) => {
		if (idx > 0) {
			content.push({ type: "hardBreak" });
		}
		if (line.length > 0) {
			content.push({ type: "text", text: line });
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
			content.push(paragraphFromText(para));
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
