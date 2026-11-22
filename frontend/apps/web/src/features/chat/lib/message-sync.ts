import type { InfiniteData } from "@tanstack/react-query";
import type { Message, MessagesList } from "@wichat/contracts/chat/message";
import {
	compareChannelSeq,
	highestChannelSeq,
	isPendingMessageId,
} from "@/features/chat/lib/message-seq";

export const CHANNEL_MESSAGES_MAX_CACHED = 3_000;
export const GAP_FILL_BURST_MESSAGES = 10_000;

export type ChannelMessagesPageParam = {
	beforeChannelSeq?: string;
	beforeCreatedAt?: string;
	beforeId?: string;
};

function normalizeMine(incoming: Message, currentUserId?: string | null): Message {
	const isMine = incoming.mine || (Boolean(currentUserId) && incoming.author_id === currentUserId);
	return { ...incoming, mine: isMine };
}

export function highestSeqInInfinite(data: InfiniteData<MessagesList> | undefined): number {
	if (!data?.pages?.length) {
		return 0;
	}
	let max = 0;
	for (const page of data.pages) {
		const pageMax = highestChannelSeq(page.messages);
		if (pageMax > max) {
			max = pageMax;
		}
	}
	return max;
}

export function applyMessageToInfinite(
	prev: InfiniteData<MessagesList> | undefined,
	incoming: Message,
	currentUserId?: string | null,
): InfiniteData<MessagesList> {
	const normalized = normalizeMine(incoming, currentUserId);

	if (!prev?.pages?.length) {
		return {
			pages: [{ messages: [normalized], has_more: false }],
			pageParams: [undefined],
		};
	}

	const pendingClientId = incoming.client_msg_id;

	let found = false;
	const nextPages = prev.pages.map((page) => {
		const idx = page.messages.findIndex(
			(m) =>
				m.id === normalized.id ||
				(pendingClientId != null && m.client_msg_id === pendingClientId) ||
				(pendingClientId != null && m.id === `pending_${pendingClientId}`),
		);
		if (idx < 0) {
			return page;
		}
		found = true;
		const updated = [...page.messages];
		const existing = page.messages[idx];
		updated[idx] = { ...normalized, mine: existing.mine || normalized.mine };
		return { ...page, messages: updated };
	});

	if (found) {
		return { ...prev, pages: nextPages };
	}

	const page0 = nextPages[0] ?? { messages: [] };
	const page0Messages = [...page0.messages, normalized].sort((a, b) => {
		const diff = compareChannelSeq(a.channel_seq, b.channel_seq);
		if (diff !== 0) {
			return diff;
		}
		return a.id.localeCompare(b.id);
	});

	return {
		...prev,
		pages: [{ ...page0, messages: page0Messages }, ...nextPages.slice(1)],
	};
}

export function replacePendingMessage(
	prev: InfiniteData<MessagesList> | undefined,
	pendingId: string,
	realMessage: Message,
	currentUserId?: string | null,
): InfiniteData<MessagesList> {
	return applyMessageToInfinite(
		removeInfiniteMessage(prev, pendingId),
		realMessage,
		currentUserId,
	);
}

function removeInfiniteMessage(
	prev: InfiniteData<MessagesList> | undefined,
	messageId: string,
): InfiniteData<MessagesList> | undefined {
	if (!prev?.pages) {
		return prev;
	}
	return {
		...prev,
		pages: prev.pages.map((page) => ({
			...page,
			messages: page.messages.filter((m) => m.id !== messageId),
		})),
	};
}

export function mergeGapFillIntoInfinite(
	prev: InfiniteData<MessagesList> | undefined,
	gapPage: MessagesList,
	currentUserId?: string | null,
): InfiniteData<MessagesList> {
	let next = prev;
	for (const msg of gapPage.messages) {
		next = applyMessageToInfinite(next, msg, currentUserId);
	}
	return trimInfiniteCache(next);
}

export function trimInfiniteCache(
	prev: InfiniteData<MessagesList> | undefined,
): InfiniteData<MessagesList> {
	if (!prev?.pages?.length) {
		return prev ?? { pages: [], pageParams: [] };
	}

	let total = 0;
	for (const page of prev.pages) {
		total += page.messages.length;
	}
	if (total <= CHANNEL_MESSAGES_MAX_CACHED) {
		return prev;
	}

	const pages = [...prev.pages];
	let excess = total - CHANNEL_MESSAGES_MAX_CACHED;
	while (excess > 0 && pages.length > 1) {
		const oldestPage = pages[pages.length - 1];
		if (oldestPage.messages.length <= excess) {
			excess -= oldestPage.messages.length;
			pages.pop();
			continue;
		}
		pages[pages.length - 1] = {
			...oldestPage,
			messages: oldestPage.messages.slice(excess),
		};
		excess = 0;
	}

	return { ...prev, pages };
}

export function oldestPageCursor(page: MessagesList): ChannelMessagesPageParam | undefined {
	if (!page.messages.length) {
		return undefined;
	}
	const oldest = page.messages[0];
	if (oldest.channel_seq) {
		return { beforeChannelSeq: oldest.channel_seq };
	}
	return {
		beforeCreatedAt: oldest.created_at,
		beforeId: oldest.id,
	};
}

export function stripPendingForDedup(messages: Message[]): Message[] {
	return messages.filter((m) => !isPendingMessageId(m.id));
}