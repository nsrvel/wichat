import type { ChannelMessage } from "@/features/chat/lib/channel-message";

/** Query key segment for TanStack Query when `services/message` lands. */
export const Q_CHANNEL_MESSAGES = "channel-messages";

export type UseChannelMessagesResult = {
	messages: ChannelMessage[];
	isPending: boolean;
	isError: boolean;
};

/** Stub until gateway exposes GET /api/v1/groups/:groupId/channels/:channelId/messages */
export function useChannelMessages(groupId: string, channelId: string): UseChannelMessagesResult {
	void [groupId, channelId];
	return {
		messages: [],
		isPending: false,
		isError: false,
	};
}
