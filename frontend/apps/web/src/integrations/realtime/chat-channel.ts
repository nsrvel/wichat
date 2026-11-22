import type {
	Message,
	MessageReactionSummary,
} from "@wichat/contracts/chat/message";
import {
	acquireGatewaySocketChannel,
	type ChannelRealtimeHandlers,
} from "@/integrations/realtime/gateway-socket";

export type ChatChannelRealtimeOptions = {
	groupId: string;
	channelId: string;
	onMessageCreated: (message: Message) => void;
	onMessageUpdated?: (message: Message) => void;
	onMessageDeleted?: (messageId: string) => void;
	onMessageReactionsUpdated?: (messageId: string, reactions: MessageReactionSummary[]) => void;
	onReconnect?: () => void;
};

/** Registers channel chat handlers on the shared gateway WebSocket. */
export function subscribeChatChannel(options: ChatChannelRealtimeOptions): () => void {
	const handlers: ChannelRealtimeHandlers = { ...options };
	return acquireGatewaySocketChannel(handlers);
}