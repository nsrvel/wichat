import type { InfiniteData } from "@tanstack/react-query";
import type { Message, MessagesList } from "@wichat/contracts/chat/message";
import { channelSeqValue } from "@/features/chat/lib/message-seq";
import type { ChannelMessagesPageParam } from "@/features/chat/lib/message-sync";

export const CHANNEL_MESSAGES_MAX_CACHED = 3_000;
export const CHANNEL_MESSAGES_JUMP_WINDOW = 120;
export const HISTORY_FETCH_MIN_INTERVAL_MS = 280;
export const HISTORY_RATE_LIMIT_COOLDOWN_MS = 60_000;

export type CacheTrimHint = "near-bottom" | "near-top" | "middle";

export type ChannelMessagesCacheMeta = {
	historyStartReached: boolean;
	oldestTrimmed: boolean;
	newestTrimmed: boolean;
	rateLimitedUntil?: number;
};

export type ChannelMessagesInfiniteData = InfiniteData<
	MessagesList,
	ChannelMessagesPageParam | undefined
> & {
	meta?: ChannelMessagesCacheMeta;
};

function defaultMeta(meta?: ChannelMessagesCacheMeta): ChannelMessagesCacheMeta {
	return {
		historyStartReached: meta?.historyStartReached ?? false,
		oldestTrimmed: meta?.oldestTrimmed ?? false,
		newestTrimmed: meta?.newestTrimmed ?? false,
		rateLimitedUntil: meta?.rateLimitedUntil,
	};
}

export function countCachedMessages(data: ChannelMessagesInfiniteData | undefined): number {
	if (!data?.pages?.length) {
		return 0;
	}
	let total = 0;
	for (const page of data.pages) {
		total += page.messages.length;
	}
	return total;
}

export function channelHistoryHasMore(data: ChannelMessagesInfiniteData | undefined): boolean {
	if (!data?.pages?.length) {
		return false;
	}
	const meta = defaultMeta(data.meta);
	if (meta.historyStartReached) {
		return false;
	}
	const lastPage = data.pages[data.pages.length - 1];
	if (lastPage.has_more !== false) {
		return true;
	}
	return meta.oldestTrimmed;
}

export function channelHasNewerBeyondCache(data: ChannelMessagesInfiniteData | undefined): boolean {
	return Boolean(data?.meta?.newestTrimmed);
}

export function isHistoryFetchRateLimited(data: ChannelMessagesInfiniteData | undefined): boolean {
	const until = data?.meta?.rateLimitedUntil;
	return typeof until === "number" && until > Date.now();
}

export function withRateLimitedUntil(
	data: ChannelMessagesInfiniteData | undefined,
	untilMs: number,
): ChannelMessagesInfiniteData {
	const base: ChannelMessagesInfiniteData = data?.pages?.length
		? data
		: { pages: [], pageParams: [] };
	return {
		...base,
		meta: {
			...defaultMeta(base.meta),
			rateLimitedUntil: untilMs,
		},
	};
}

export function clearRateLimitedIfExpired(
	data: ChannelMessagesInfiniteData | undefined,
): ChannelMessagesInfiniteData | undefined {
	if (!data || !isHistoryFetchRateLimited(data)) {
		return data;
	}
	return {
		...data,
		meta: {
			...defaultMeta(data.meta),
			rateLimitedUntil: undefined,
		},
	};
}

export function recordHistoryPageMeta(
	prev: ChannelMessagesInfiniteData | undefined,
	page: MessagesList,
): ChannelMessagesInfiniteData | undefined {
	if (!prev) {
		return prev;
	}
	const meta = defaultMeta(prev.meta);
	if (page.has_more === false) {
		meta.historyStartReached = true;
		meta.oldestTrimmed = false;
	}
	return { ...prev, meta };
}

function trimFromOldest(
	pages: MessagesList[],
	pageParams: (ChannelMessagesPageParam | undefined)[],
	excess: number,
	meta: ChannelMessagesCacheMeta,
): { pages: MessagesList[]; pageParams: (ChannelMessagesPageParam | undefined)[]; excess: number } {
	let remaining = excess;
	const nextPages = [...pages];
	const nextParams = [...pageParams];

	while (remaining > 0 && nextPages.length > 1) {
		const oldestPage = nextPages[nextPages.length - 1];
		if (oldestPage.messages.length <= remaining) {
			remaining -= oldestPage.messages.length;
			nextPages.pop();
			nextParams.pop();
			meta.oldestTrimmed = true;
			continue;
		}
		nextPages[nextPages.length - 1] = {
			...oldestPage,
			messages: oldestPage.messages.slice(remaining),
			has_more: true,
		};
		meta.oldestTrimmed = true;
		remaining = 0;
	}

	return { pages: nextPages, pageParams: nextParams, excess: remaining };
}

function trimFromNewest(
	pages: MessagesList[],
	pageParams: (ChannelMessagesPageParam | undefined)[],
	excess: number,
	meta: ChannelMessagesCacheMeta,
): { pages: MessagesList[]; pageParams: (ChannelMessagesPageParam | undefined)[]; excess: number } {
	let remaining = excess;
	const nextPages = [...pages];
	const nextParams = [...pageParams];

	while (remaining > 0 && nextPages.length > 0) {
		const newestPage = nextPages[0];
		if (newestPage.messages.length <= remaining) {
			remaining -= newestPage.messages.length;
			nextPages.shift();
			nextParams.shift();
			meta.newestTrimmed = true;
			continue;
		}
		nextPages[0] = {
			...newestPage,
			messages: newestPage.messages.slice(0, newestPage.messages.length - remaining),
			has_more: true,
		};
		meta.newestTrimmed = true;
		remaining = 0;
	}

	return { pages: nextPages, pageParams: nextParams, excess: remaining };
}

/**
 * Enforces CHANNEL_MESSAGES_MAX_CACHED while trimming away from the user's scroll anchor.
 */
export function applyBoundedMessageCache(
	prev: ChannelMessagesInfiniteData | undefined,
	hint: CacheTrimHint = "near-bottom",
	maxMessages: number = CHANNEL_MESSAGES_MAX_CACHED,
): ChannelMessagesInfiniteData {
	if (!prev?.pages?.length) {
		return prev ?? { pages: [], pageParams: [] };
	}

	const total = countCachedMessages(prev);
	if (total <= maxMessages) {
		return prev;
	}

	const meta = defaultMeta(prev.meta);
	let pages = prev.pages.map((p) => ({ ...p, messages: [...p.messages] }));
	let pageParams = [...(prev.pageParams ?? [])];
	let excess = total - maxMessages;

	if (hint === "near-bottom") {
		({ pages, pageParams, excess } = trimFromOldest(pages, pageParams, excess, meta));
		if (excess > 0) {
			({ pages, pageParams } = trimFromNewest(pages, pageParams, excess, meta));
		}
	} else if (hint === "near-top") {
		({ pages, pageParams, excess } = trimFromNewest(pages, pageParams, excess, meta));
		if (excess > 0) {
			({ pages, pageParams } = trimFromOldest(pages, pageParams, excess, meta));
		}
	} else {
		const fromOldest = Math.ceil(excess / 2);
		({ pages, pageParams, excess } = trimFromOldest(pages, pageParams, fromOldest, meta));
		({ pages, pageParams } = trimFromNewest(pages, pageParams, excess, meta));
	}

	if (meta.oldestTrimmed && pages.length > 0 && !meta.historyStartReached) {
		const tail = pages[pages.length - 1];
		if (tail.has_more === false) {
			pages[pages.length - 1] = { ...tail, has_more: true };
		}
	}

	return {
		...prev,
		pages,
		pageParams,
		meta,
	};
}

export function appendOlderChannelPage(
	prev: ChannelMessagesInfiniteData | undefined,
	page: MessagesList,
	pageParam: ChannelMessagesPageParam | undefined,
): ChannelMessagesInfiniteData {
	const base: ChannelMessagesInfiniteData = prev?.pages?.length
		? prev
		: { pages: [], pageParams: [] };
	const next: ChannelMessagesInfiniteData = {
		...base,
		pages: [...base.pages, page],
		pageParams: [...(base.pageParams ?? []), pageParam],
		meta: defaultMeta(base.meta),
	};
	if (page.has_more === false) {
		next.meta = { ...defaultMeta(next.meta), historyStartReached: true, oldestTrimmed: false };
	}
	return next;
}

export function mergeSortedMessagesUnique(existing: Message[], incoming: Message[]): Message[] {
	const byId = new Map<string, Message>();
	for (const msg of existing) {
		byId.set(msg.id, msg);
	}
	for (const msg of incoming) {
		byId.set(msg.id, msg);
	}
	return [...byId.values()].sort((a, b) => {
		const diff = channelSeqValue(a.channel_seq) - channelSeqValue(b.channel_seq);
		if (diff !== 0) {
			return diff;
		}
		return a.created_at.localeCompare(b.created_at) || a.id.localeCompare(b.id);
	});
}

export function installContextMessageWindow(
	_messages: Message[],
	hasOlder: boolean,
	hasNewer: boolean,
): ChannelMessagesInfiniteData {
	const sorted = [..._messages].sort(
		(a, b) => channelSeqValue(a.channel_seq) - channelSeqValue(b.channel_seq),
	);
	return applyBoundedMessageCache(
		{
			pages: [{ messages: sorted, has_more: !hasOlder }],
			pageParams: [undefined],
			meta: {
				historyStartReached: !hasOlder,
				oldestTrimmed: hasOlder,
				newestTrimmed: hasNewer,
			},
		},
		"middle",
	);
}

export function findMessageSeqInCache(
	data: ChannelMessagesInfiniteData | undefined,
	messageId: string,
): number {
	if (!data?.pages?.length || !messageId) {
		return 0;
	}
	for (const page of data.pages) {
		for (const msg of page.messages) {
			if (msg.id === messageId) {
				return channelSeqValue(msg.channel_seq);
			}
		}
	}
	return 0;
}

export function oldestPageCursor(page: MessagesList): ChannelMessagesPageParam | undefined {
	if (!page.messages.length) {
		return undefined;
	}
	const oldest = page.messages[0];
	const seq = channelSeqValue(oldest.channel_seq);
	if (seq <= 0) {
		return undefined;
	}
	return { beforeChannelSeq: seq };
}

export function oldestCacheCursor(
	data: ChannelMessagesInfiniteData | undefined,
): ChannelMessagesPageParam | undefined {
	if (!data?.pages?.length) {
		return undefined;
	}
	const lastPage = data.pages[data.pages.length - 1];
	return oldestPageCursor(lastPage);
}

/** Back-compat wrapper used by gap-fill and legacy callers. */
export function trimInfiniteCache(
	prev: InfiniteData<MessagesList> | undefined,
	hint: CacheTrimHint = "near-bottom",
): ChannelMessagesInfiniteData {
	return applyBoundedMessageCache(prev as ChannelMessagesInfiniteData, hint);
}
