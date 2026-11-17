import { infiniteQueryOptions, queryOptions, type QueryClient } from "@tanstack/react-query";
import { isPublicEntityId } from "@wichat/contracts/common";
import { Q_CHANNEL_MESSAGES, Q_PINNED_MESSAGES } from "@/services/message/message.key";
import {
	CHANNEL_MESSAGES_PAGE_SIZE,
	listChannelMessages,
	listPinnedChannelMessages,
} from "@/services/message/message.service";

export function channelMessagesQueryOpts(groupId: string, channelId: string) {
	return infiniteQueryOptions({
		queryKey: [Q_CHANNEL_MESSAGES, groupId, channelId],
		initialPageParam: undefined as string | undefined,
		queryFn: ({ pageParam }) =>
			listChannelMessages(groupId, channelId, {
				limit: CHANNEL_MESSAGES_PAGE_SIZE,
				beforeCreatedAt: pageParam,
			}),
		getNextPageParam: (lastPage) => {
			if (!lastPage?.messages || lastPage.messages.length < CHANNEL_MESSAGES_PAGE_SIZE) {
				return undefined;
			}
			return lastPage.messages[0]?.created_at;
		},
		maxPages: 10,
		enabled: isPublicEntityId(groupId) && isPublicEntityId(channelId),
		staleTime: 30_000,
		refetchOnWindowFocus: false,
	});
}

export function prefetchChannelMessages(
	queryClient: QueryClient,
	groupId: string,
	channelId: string,
) {
	if (!isPublicEntityId(groupId) || !isPublicEntityId(channelId)) {
		return;
	}
	void queryClient.prefetchInfiniteQuery(channelMessagesQueryOpts(groupId, channelId));
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
