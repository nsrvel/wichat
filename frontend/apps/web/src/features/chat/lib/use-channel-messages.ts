import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import type { Message } from "@wichat/contracts/chat/message";
import { useCallback, useLayoutEffect, useMemo, useRef } from "react";
import {
	flattenInfiniteChannelMessages,
	type ChannelMessage,
} from "@/features/chat/lib/channel-message";
import { formatMessageTimeLabel } from "@/features/chat/lib/format-message-time";
import { useTimeFormatPreference, type TimeFormatPreference } from "@/lib/time-format";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import { channelMessagesQueryOpts } from "@/services/message/message.query";

export type UseChannelMessagesResult = {
	messages: ChannelMessage[];
	isPending: boolean;
	isFetching: boolean;
	isError: boolean;
	hasOlderMessages: boolean;
	isFetchingOlderMessages: boolean;
	fetchOlderMessages: () => void;
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

export function useChannelMessages(groupId: string, channelId: string): UseChannelMessagesResult {
	const {
		data,
		isPending,
		isFetching,
		isError,
		hasNextPage,
		isFetchingNextPage,
		fetchNextPage,
		refetch: queryRefetch,
	} = useInfiniteQuery(channelMessagesQueryOpts(groupId, channelId));
	const { data: sessionUser } = useQuery(sessionQueryOpts());
	const timeFormat = useTimeFormatPreference();
	const currentUserId = sessionUser?.id ?? "";

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

	const paginationStateRef = useRef({ hasNextPage, isFetchingNextPage });
	useLayoutEffect(() => {
		paginationStateRef.current = { hasNextPage, isFetchingNextPage };
	}, [hasNextPage, isFetchingNextPage]);

	const fetchOlderMessages = useCallback(() => {
		const { hasNextPage: canFetch, isFetchingNextPage: busy } = paginationStateRef.current;
		if (canFetch && !busy) {
			void fetchNextPage();
		}
	}, [fetchNextPage]);

	const refetch = useCallback(() => {
		void queryRefetch();
	}, [queryRefetch]);

	return {
		messages,
		isPending,
		isFetching,
		isError,
		hasOlderMessages: Boolean(hasNextPage),
		isFetchingOlderMessages: isFetchingNextPage,
		fetchOlderMessages,
		refetch,
	};
}
