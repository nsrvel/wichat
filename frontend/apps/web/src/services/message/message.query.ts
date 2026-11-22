import { infiniteQueryOptions, queryOptions } from "@tanstack/react-query";
import { isPublicEntityId } from "@wichat/contracts/common";
import {
	oldestPageCursor,
	type ChannelMessagesPageParam,
} from "@/features/chat/lib/message-sync";
import { Q_CHANNEL_MESSAGES, Q_PINNED_MESSAGES } from "@/services/message/message.key";
import {
	CHANNEL_MESSAGES_PAGE_SIZE,
	listChannelMessages,
	listPinnedChannelMessages,
} from "@/services/message/message.service";

export type { ChannelMessagesPageParam };

export function channelMessagesQueryOpts(groupId: string, channelId: string) {
	return infiniteQueryOptions({
		queryKey: [Q_CHANNEL_MESSAGES, groupId, channelId],
		initialPageParam: undefined as ChannelMessagesPageParam | undefined,
		queryFn: ({ pageParam, signal }) =>
			listChannelMessages(groupId, channelId, {
				limit: CHANNEL_MESSAGES_PAGE_SIZE,
				beforeCreatedAt: pageParam?.beforeCreatedAt,
				beforeId: pageParam?.beforeId,
				beforeChannelSeq: pageParam?.beforeChannelSeq,
				signal,
			}),
		getNextPageParam: (lastPage) => {
			if (!lastPage?.messages || lastPage.messages.length === 0) {
				return undefined;
			}
			const hasMore =
				typeof lastPage.has_more === "boolean"
					? lastPage.has_more
					: lastPage.messages.length >= CHANNEL_MESSAGES_PAGE_SIZE;
			if (!hasMore) {
				return undefined;
			}
			return oldestPageCursor(lastPage);
		},
		enabled: isPublicEntityId(groupId) && isPublicEntityId(channelId),
		staleTime: 30_000,
		refetchOnWindowFocus: false,
	});
}

export function pinnedChannelMessagesQueryOpts(groupId: string, channelId: string) {
	return queryOptions({
		queryKey: [Q_PINNED_MESSAGES, groupId, channelId],
		queryFn: () => listPinnedChannelMessages(groupId, channelId),
		enabled: isPublicEntityId(groupId) && isPublicEntityId(channelId),
		staleTime: 60_000,
		refetchOnWindowFocus: false,
	});
}
