import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import type { Message } from "@wichat/contracts/chat/message";
import { useCallback, useMemo } from "react";
import {
	flattenInfiniteChannelMessages,
	type ChannelMessage,
} from "@/features/chat/lib/channel-message";
import { formatMessageTimeLabel } from "@/features/chat/lib/format-message-time";
import { useTimeFormatPreference } from "@/lib/time-format";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import { channelMessagesQueryOpts } from "@/services/message/message.query";

export type UseChannelMessagesResult = {
	messages: ChannelMessage[];
	isPending: boolean;
	isError: boolean;
	hasOlderMessages: boolean;
	isFetchingOlderMessages: boolean;
	fetchOlderMessages: () => void;
	refetch: () => void;
};

type CachedEntry = {
	userId: string;
	timeFormat: string;
	value: ChannelMessage;
};

const messageViewCache = new WeakMap<Message, CachedEntry>();

export function useChannelMessages(groupId: string, channelId: string): UseChannelMessagesResult {
	const query = useInfiniteQuery(channelMessagesQueryOpts(groupId, channelId));
	const { data: sessionUser } = useQuery(sessionQueryOpts());
	const timeFormat = useTimeFormatPreference();
	const currentUserId = sessionUser?.id ?? "";

	const messages = useMemo<ChannelMessage[]>(() => {
		const rawMessages = flattenInfiniteChannelMessages(query.data);

		return rawMessages.map((msg) => {
			const hit = messageViewCache.get(msg);
			if (hit && hit.userId === currentUserId && hit.timeFormat === timeFormat) {
				return hit.value;
			}
			const mapped: ChannelMessage = {
				id: msg.id,
				authorId: msg.author_id,
				authorDisplayName: msg.author_display_name,
				authorAvatarUrl: msg.author_avatar_url ?? "",
				createdAt: msg.created_at,
				editedAt: msg.edited_at || undefined,
				isPinned: msg.is_pinned ?? false,
				replyTo: msg.reply_to ?? null,
				bodyPlain: msg.body,
				mine: msg.mine || (Boolean(currentUserId) && msg.author_id === currentUserId),
				reactions: (msg.reactions ?? []).map((r) => ({
					emoji: r.emoji,
					count: r.count,
					active: Boolean(currentUserId) && r.user_ids.includes(currentUserId),
					userIds: r.user_ids,
				})),
				timeLabel: formatMessageTimeLabel(msg.created_at, timeFormat),
			};
			messageViewCache.set(msg, { userId: currentUserId, timeFormat, value: mapped });
			return mapped;
		});
	}, [query.data, currentUserId, timeFormat]);

	const fetchOlderMessages = useCallback(() => {
		if (query.hasNextPage && !query.isFetchingNextPage) {
			void query.fetchNextPage();
		}
	}, [query]);

	const refetch = useCallback(() => {
		void query.refetch();
	}, [query]);

	return {
		messages,
		isPending: query.isPending,
		isError: query.isError,
		hasOlderMessages: Boolean(query.hasNextPage),
		isFetchingOlderMessages: query.isFetchingNextPage,
		fetchOlderMessages,
		refetch,
	};
}
