import { queryOptions } from "@tanstack/react-query";
import { isPublicEntityId } from "@wichat/contracts";
import { Q_CHANNEL_MESSAGES } from "@/services/message/message.key";
import { listChannelMessages } from "@/services/message/message.service";

export function channelMessagesQueryOpts(groupId: string, channelId: string) {
	return queryOptions({
		queryKey: [Q_CHANNEL_MESSAGES, groupId, channelId],
		queryFn: () => listChannelMessages(groupId, channelId),
		enabled: isPublicEntityId(groupId) && isPublicEntityId(channelId),
		staleTime: 30_000,
		refetchOnWindowFocus: true,
	});
}
