import { useQuery } from "@tanstack/react-query";
import type { ChannelMessage } from "@/features/chat/lib/channel-message";
import { formatMessageTimeLabel } from "@/features/chat/lib/format-message-time";
import { channelMessagesQueryOpts } from "@/services/message/message.query";

export type UseChannelMessagesResult = {
	messages: ChannelMessage[];
	isPending: boolean;
	isError: boolean;
};

export function useChannelMessages(groupId: string, channelId: string): UseChannelMessagesResult {
	const query = useQuery(channelMessagesQueryOpts(groupId, channelId));

	const messages: ChannelMessage[] = (query.data?.messages ?? []).map((msg) => ({
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
	};
}
