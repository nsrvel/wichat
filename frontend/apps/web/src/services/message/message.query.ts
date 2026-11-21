import { infiniteQueryOptions, queryOptions } from "@tanstack/react-query";
import { isPublicEntityId } from "@wichat/contracts/common";
import { Q_CHANNEL_MESSAGES, Q_PINNED_MESSAGES } from "@/services/message/message.key";
import {
	CHANNEL_MESSAGES_PAGE_SIZE,
	listChannelMessages,
	listPinnedChannelMessages,
} from "@/services/message/message.service";

export type ChannelMessagesPageParam = {
	beforeCreatedAt?: string;
	beforeId?: string;
};

export function channelMessagesQueryOpts(groupId: string, channelId: string) {
	return infiniteQueryOptions({
		queryKey: [Q_CHANNEL_MESSAGES, groupId, channelId],
		initialPageParam: undefined as ChannelMessagesPageParam | undefined,
		queryFn: ({ pageParam }) =>
			listChannelMessages(groupId, channelId, {
				limit: CHANNEL_MESSAGES_PAGE_SIZE,
				beforeCreatedAt: pageParam?.beforeCreatedAt,
				beforeId: pageParam?.beforeId,
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
			const oldest = lastPage.messages[0];
			return oldest ? { beforeCreatedAt: oldest.created_at, beforeId: oldest.id } : undefined;
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
