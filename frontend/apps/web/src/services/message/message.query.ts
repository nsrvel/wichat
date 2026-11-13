import { infiniteQueryOptions } from "@tanstack/react-query";
import { isPublicEntityId } from "@wichat/contracts";
import { Q_CHANNEL_MESSAGES } from "@/services/message/message.key";
import {
	CHANNEL_MESSAGES_PAGE_SIZE,
	listChannelMessages,
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
		enabled: isPublicEntityId(groupId) && isPublicEntityId(channelId),
		staleTime: 30_000,
		refetchOnWindowFocus: true,
	});
}
