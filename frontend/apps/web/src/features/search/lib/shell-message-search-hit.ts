import type { InfiniteData } from "@tanstack/react-query";
import type { Message, MessagesList } from "@wichat/contracts/chat/message";
import { flattenInfiniteChannelMessages } from "@/features/chat/lib/channel-message";

export type ShellMessageSearchHit = {
	groupId: string;
	channelId: string;
	messageId: string;
	bodyPreview: string;
	authorDisplayName: string;
	createdAt: string;
};

const MAX_HITS = 8;

function normalizeBody(body: string | undefined): string {
	const raw = (body ?? "").trim();
	return raw.replace(/\s+/g, " ");
}

/** Offline fallback: scan loaded channel message cache (current channel only). */
export function spotlightChannelMessages(
	data: InfiniteData<MessagesList> | undefined,
	query: string,
	groupId: string,
	channelId: string,
): ShellMessageSearchHit[] {
	const q = query.trim().toLowerCase();
	if (!q || !data || !groupId || !channelId) {
		return [];
	}

	const messages = flattenInfiniteChannelMessages(data);
	const hits: ShellMessageSearchHit[] = [];

	for (let i = messages.length - 1; i >= 0; i -= 1) {
		const msg: Message | undefined = messages[i];
		if (!msg) {
			continue;
		}
		const body = normalizeBody(msg.body);
		if (!body.toLowerCase().includes(q)) {
			continue;
		}
		hits.push({
			groupId,
			channelId,
			messageId: msg.id,
			bodyPreview: body.length > 140 ? `${body.slice(0, 137)}…` : body,
			authorDisplayName: msg.author_display_name || "",
			createdAt: msg.created_at,
		});
		if (hits.length >= MAX_HITS) {
			break;
		}
	}

	return hits;
}