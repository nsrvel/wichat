import { ApiError } from "@wichat/api-client/api-error";
import { useInfiniteQuery, useQuery, useQueryClient } from "@tanstack/react-query";
import type { Message } from "@wichat/contracts/chat/message";
import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import {
	flattenInfiniteChannelMessages,
	type ChannelMessage,
} from "@/features/chat/lib/channel-message";
import { useChatCacheTrimHintRef } from "@/features/chat/lib/chat-cache-trim-hint";
import { formatMessageTimeLabel } from "@/features/chat/lib/format-message-time";
import {
	appendOlderChannelPage,
	applyBoundedMessageCache,
	channelHistoryHasMore,
	type ChannelMessagesInfiniteData,
	clearRateLimitedIfExpired,
	countCachedMessages,
	HISTORY_FETCH_MIN_INTERVAL_MS,
	HISTORY_RATE_LIMIT_COOLDOWN_MS,
	installContextMessageWindow,
	isHistoryFetchRateLimited,
	oldestCacheCursor,
	recordHistoryPageMeta,
	withRateLimitedUntil,
} from "@/features/chat/lib/message-cache-window";
import { useTimeFormatPreference, type TimeFormatPreference } from "@/lib/time-format";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import { Q_CHANNEL_MESSAGES } from "@/services/message/message.key";
import { channelMessagesQueryOpts } from "@/services/message/message.query";
import {
	CHANNEL_MESSAGES_PAGE_SIZE,
	listChannelMessages,
	listChannelMessagesAroundSeq,
} from "@/services/message/message.service";

export type UseChannelMessagesResult = {
	messages: ChannelMessage[];
	isPending: boolean;
	isFetching: boolean;
	isError: boolean;
	hasOlderMessages: boolean;
	isFetchingOlderMessages: boolean;
	isHistoryLoadPaused: boolean;
	fetchOlderMessages: () => void;
	loadContextAroundSeq: (centerSeq: number) => Promise<boolean>;
	refetch: () => void;
};

const EMPTY_REACTIONS: NonNullable<ChannelMessage["reactions"]> = [];

let cachedContextKey = "";
let messageWeakCache = new WeakMap<Message, ChannelMessage>();

function getMessageWeakCache(
	currentUserId: string,
	timeFormat: TimeFormatPreference,
): WeakMap<Message, ChannelMessage> {
	const nextKey = `${currentUserId}|${timeFormat}`;
	if (cachedContextKey !== nextKey) {
		cachedContextKey = nextKey;
		messageWeakCache = new WeakMap<Message, ChannelMessage>();
	}
	return messageWeakCache;
}

function computeLocalDayKey(date: Date): string {
	if (Number.isNaN(date.getTime())) {
		return "unknown";
	}
	return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

function toChannelMessage(
	msg: Message,
	currentUserId: string,
	timeFormat: TimeFormatPreference,
): ChannelMessage {
	const createdDate = new Date(msg.created_at);
	const createdAtMs = createdDate.getTime();
	const rawReactions = msg.reactions ?? [];
	return {
		id: msg.id,
		clientMsgId: msg.client_msg_id,
		authorId: msg.author_id,
		authorDisplayName: msg.author_display_name,
		authorAvatarUrl: msg.author_avatar_url ?? "",
		createdAt: msg.created_at,
		createdAtMs: Number.isNaN(createdAtMs) ? undefined : createdAtMs,
		dayKey: computeLocalDayKey(createdDate),
		editedAt: msg.edited_at || undefined,
		isPinned: msg.is_pinned ?? false,
		replyTo: msg.reply_to ?? null,
		bodyPlain: msg.body,
		mine: msg.mine || (Boolean(currentUserId) && msg.author_id === currentUserId),
		reactions:
			rawReactions.length === 0
				? EMPTY_REACTIONS
				: rawReactions.map((r) => ({
						emoji: r.emoji,
						count: r.count,
						active: Boolean(currentUserId) && r.user_ids.includes(currentUserId),
						userIds: r.user_ids,
					})),
		timeLabel: formatMessageTimeLabel(msg.created_at, timeFormat),
	};
}

function isRateLimitError(error: unknown): boolean {
	return error instanceof ApiError && error.code === "error.rate_limited";
}

export function useChannelMessages(groupId: string, channelId: string): UseChannelMessagesResult {
	const queryClient = useQueryClient();
	const trimHintRef = useChatCacheTrimHintRef();
	const queryKey = useMemo(() => [Q_CHANNEL_MESSAGES, groupId, channelId], [groupId, channelId]);

	const {
		data: rawData,
		isPending,
		isFetching,
		isError,
		hasNextPage,
		isFetchingNextPage,
		fetchNextPage,
		refetch: queryRefetch,
	} = useInfiniteQuery(channelMessagesQueryOpts(groupId, channelId));

	const data = rawData as ChannelMessagesInfiniteData | undefined;

	const { data: sessionUser } = useQuery(sessionQueryOpts());
	const timeFormat = useTimeFormatPreference();
	const currentUserId = sessionUser?.id ?? "";

	useLayoutEffect(() => {
		if (!data?.pages?.length) {
			return;
		}
		let working = clearRateLimitedIfExpired(data) ?? data;
		const tail = working.pages[working.pages.length - 1];
		working = recordHistoryPageMeta(working, tail) ?? working;
		const beforeCount = countCachedMessages(working);
		const bounded = applyBoundedMessageCache(working, trimHintRef.current);
		const afterCount = countCachedMessages(bounded);
		if (
			afterCount < beforeCount ||
			bounded.meta?.historyStartReached !== working.meta?.historyStartReached ||
			bounded.meta?.oldestTrimmed !== working.meta?.oldestTrimmed ||
			bounded.meta?.newestTrimmed !== working.meta?.newestTrimmed ||
			bounded.meta?.rateLimitedUntil !== working.meta?.rateLimitedUntil
		) {
			queryClient.setQueryData(queryKey, bounded);
		}
	}, [data, queryClient, queryKey, trimHintRef]);

	const messages: ChannelMessage[] = useMemo(() => {
		const cache = getMessageWeakCache(currentUserId, timeFormat);
		const rawMessages = flattenInfiniteChannelMessages(data);
		return rawMessages.map((msg) => {
			const hit = cache.get(msg);
			if (hit) {
				return hit;
			}
			const mapped = toChannelMessage(msg, currentUserId, timeFormat);
			cache.set(msg, mapped);
			return mapped;
		});
	}, [data, currentUserId, timeFormat]);

	const hasOlderFromCache = channelHistoryHasMore(data);
	const isHistoryLoadPaused = isHistoryFetchRateLimited(data);
	const [manualOlderFetching, setManualOlderFetching] = useState(false);
	const lastOlderFetchAtRef = useRef(0);

	useEffect(() => {
		if (!isHistoryLoadPaused) {
			return;
		}
		const timer = setInterval(() => {
			queryClient.setQueryData<ChannelMessagesInfiniteData>(queryKey, (prev) => {
				const cleared = clearRateLimitedIfExpired(prev);
				return cleared ?? prev;
			});
		}, 5_000);
		return () => clearInterval(timer);
	}, [isHistoryLoadPaused, queryClient, queryKey]);

	const paginationStateRef = useRef({
		hasNextPage,
		hasOlderFromCache,
		isFetchingNextPage,
		isHistoryLoadPaused,
		manualOlderFetching,
	});
	useLayoutEffect(() => {
		paginationStateRef.current = {
			hasNextPage,
			hasOlderFromCache,
			isFetchingNextPage,
			isHistoryLoadPaused,
			manualOlderFetching,
		};
	}, [hasNextPage, hasOlderFromCache, isFetchingNextPage, isHistoryLoadPaused, manualOlderFetching]);

	const fetchOlderMessages = useCallback(() => {
		const state = paginationStateRef.current;
		if (state.isHistoryLoadPaused || state.isFetchingNextPage || state.manualOlderFetching) {
			return;
		}
		if (!state.hasOlderFromCache && !state.hasNextPage) {
			return;
		}

		const now = Date.now();
		if (now - lastOlderFetchAtRef.current < HISTORY_FETCH_MIN_INTERVAL_MS) {
			return;
		}
		lastOlderFetchAtRef.current = now;

		const runManualOlder = async () => {
			const cursor = oldestCacheCursor(
				queryClient.getQueryData<ChannelMessagesInfiniteData>(queryKey),
			);
			if (!cursor?.beforeChannelSeq) {
				return;
			}
			setManualOlderFetching(true);
			try {
				const page = await listChannelMessages(groupId, channelId, {
					limit: CHANNEL_MESSAGES_PAGE_SIZE,
					beforeChannelSeq: cursor.beforeChannelSeq,
				});
				queryClient.setQueryData<ChannelMessagesInfiniteData>(queryKey, (prev) => {
					let next = appendOlderChannelPage(prev, page, cursor);
					next = recordHistoryPageMeta(next, page) ?? next;
					return applyBoundedMessageCache(next, trimHintRef.current);
				});
			} catch (error) {
				if (isRateLimitError(error)) {
					queryClient.setQueryData<ChannelMessagesInfiniteData>(queryKey, (prev) =>
						withRateLimitedUntil(prev, Date.now() + HISTORY_RATE_LIMIT_COOLDOWN_MS),
					);
				}
			} finally {
				setManualOlderFetching(false);
			}
		};

		if (state.hasNextPage) {
			void fetchNextPage().catch((error: unknown) => {
				if (isRateLimitError(error)) {
					queryClient.setQueryData<ChannelMessagesInfiniteData>(queryKey, (prev) =>
						withRateLimitedUntil(prev, Date.now() + HISTORY_RATE_LIMIT_COOLDOWN_MS),
					);
					return;
				}
				throw error;
			});
			return;
		}

		void runManualOlder();
	}, [channelId, fetchNextPage, groupId, queryClient, queryKey, trimHintRef]);

	const loadContextAroundSeq = useCallback(
		async (centerSeq: number): Promise<boolean> => {
			if (centerSeq <= 0) {
				return false;
			}
			try {
				const window = await listChannelMessagesAroundSeq(groupId, channelId, centerSeq);
				if (!window.messages.length) {
					return false;
				}
				queryClient.setQueryData<ChannelMessagesInfiniteData>(queryKey, () =>
					applyBoundedMessageCache(
						installContextMessageWindow(
							window.messages,
							window.hasOlder,
							window.hasNewer,
						),
						"middle",
					),
				);
				return true;
			} catch {
				return false;
			}
		},
		[channelId, groupId, queryClient, queryKey],
	);

	const refetch = useCallback(() => {
		void queryRefetch();
	}, [queryRefetch]);

	return {
		messages,
		isPending,
		isFetching,
		isError,
		hasOlderMessages: hasOlderFromCache || Boolean(hasNextPage),
		isFetchingOlderMessages: isFetchingNextPage || manualOlderFetching,
		isHistoryLoadPaused,
		fetchOlderMessages,
		loadContextAroundSeq,
		refetch,
	};
}