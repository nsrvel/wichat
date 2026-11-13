import { useInfiniteQuery } from "@tanstack/react-query";
import {
	flattenInfiniteChannelMessages,
	type ChannelMessage,
} from "@/features/chat/lib/channel-message";
import { formatMessageTimeLabel } from "@/features/chat/lib/format-message-time";
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

export function useChannelMessages(groupId: string, channelId: string): UseChannelMessagesResult {
	const query = useInfiniteQuery(channelMessagesQueryOpts(groupId, channelId));

	const rawMessages = flattenInfiniteChannelMessages(query.data);
	const messages: ChannelMessage[] = rawMessages.map((msg) => ({
		id: msg.id,
		authorId: msg.author_id,
		authorDisplayName: msg.author_display_name,
		authorAvatarUrl: "",
		createdAt: msg.created_at,
		bodyPlain: msg.body,
		mine: msg.mine,
		timeLabel: formatMessageTimeLabel(msg.created_at),
	}));

	return {
		messages,
		isPending: query.isPending,
		isError: query.isError,
		hasOlderMessages: Boolean(query.hasNextPage),
		isFetchingOlderMessages: query.isFetchingNextPage,
		fetchOlderMessages: () => {
			if (query.hasNextPage && !query.isFetchingNextPage) {
				void query.fetchNextPage();
			}
		},
		refetch: () => {
			void query.refetch();
		},
	};
}
