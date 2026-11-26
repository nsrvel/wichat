import type { JSONContent } from "@tiptap/react";
import { extractMentionPublicUserIds } from "@/features/chat/lib/mention-markdown";

function collectFromNodes(nodes: JSONContent[] | undefined, out: Set<string>): void {
	if (!nodes?.length) {
		return;
	}
	for (const node of nodes) {
		if (node.type === "userMention") {
			const id = typeof node.attrs?.id === "string" ? node.attrs.id.trim() : "";
			if (id) {
				out.add(id);
			}
		}
		collectFromNodes(node.content, out);
	}
}

/** Unique public user ids from composer doc and/or serialized markdown body. */
export function extractMentionUserIdsFromComposer(
	doc: JSONContent | undefined,
	bodyMarkdown: string,
): string[] {
	const out = new Set<string>();
	collectFromNodes(doc?.content, out);
	for (const id of extractMentionPublicUserIds(bodyMarkdown)) {
		out.add(id);
	}
	return [...out];
}