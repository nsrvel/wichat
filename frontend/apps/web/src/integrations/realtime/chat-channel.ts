import {
	RealtimeChatEventSchema,
	type Message,
	type MessageReactionSummary,
} from "@wichat/contracts/chat/message";
import { ensureAccessToken, getAccessToken } from "@/integrations/api-client";
import { onGatewayRecovered } from "@/integrations/realtime/gateway-connection";
import { env } from "@/lib/env";

export type ChatChannelRealtimeOptions = {
	groupId: string;
	channelId: string;
	onMessageCreated: (message: Message) => void;
	onMessageUpdated?: (message: Message) => void;
	onMessageDeleted?: (messageId: string) => void;
	onMessageReactionsUpdated?: (messageId: string, reactions: MessageReactionSummary[]) => void;
	onReconnect?: () => void;
};

function realtimeWebSocketURL(accessToken: string): string {
	const base = env.NEXT_PUBLIC_API_URL.replace(/\/+$/, "").replace(/^http/, "ws");
	const pathname = base.endsWith("/api/v1") ? `${base}/realtime` : `${base}/api/v1/realtime`;
	const url = new URL(pathname);
	url.searchParams.set("access_token", accessToken);
	return url.toString();
}

/** Opens a WebSocket, subscribes to one channel, and forwards chat lifecycle events. */
export function subscribeChatChannel(options: ChatChannelRealtimeOptions): () => void {
	let ws: WebSocket | null = null;
	let closed = false;
	let hasConnectedOnce = false;
	let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
	let reconnectAttempts = 0;

	const scheduleReconnect = () => {
		if (closed || reconnectTimer !== undefined) {
			return;
		}
		const delay = Math.min(1000 * 2 ** reconnectAttempts, 10000);
		reconnectAttempts++;
		reconnectTimer = setTimeout(() => {
			reconnectTimer = undefined;
			void connect();
		}, delay);
	};

	const connect = async () => {
		if (closed) {
			return;
		}
		let token = getAccessToken();
		if (!token) {
			token = await ensureAccessToken();
		}
		if (closed) {
			return;
		}
		if (!token) {
			scheduleReconnect();
			return;
		}

		try {
			ws = new WebSocket(realtimeWebSocketURL(token));
		} catch {
			scheduleReconnect();
			return;
		}

		ws.addEventListener("open", () => {
			reconnectAttempts = 0;
			if (hasConnectedOnce) {
				options.onReconnect?.();
			}
			hasConnectedOnce = true;
			ws?.send(
				JSON.stringify({
					type: "subscribe",
					group_id: options.groupId,
					channel_id: options.channelId,
				}),
			);
		});
		ws.addEventListener("message", (event) => {
			try {
				const parsed = RealtimeChatEventSchema.parse(JSON.parse(String(event.data)));
				if (parsed.type === "message.created") {
					options.onMessageCreated(parsed.message);
				} else if (parsed.type === "message.updated") {
					options.onMessageUpdated?.(parsed.message);
				} else if (parsed.type === "message.deleted") {
					options.onMessageDeleted?.(parsed.message_id);
				} else if (parsed.type === "message.reactions.updated") {
					options.onMessageReactionsUpdated?.(parsed.message_id, parsed.reactions);
				}
			} catch {
				// ignore malformed or non-chat fanout frames
			}
		});
		ws.addEventListener("close", () => {
			ws = null;
			if (!closed) {
				scheduleReconnect();
			}
		});
	};

	const unsubscribeRecovered = onGatewayRecovered(() => {
		if (closed) {
			return;
		}
		if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
			return;
		}
		if (reconnectTimer !== undefined) {
			clearTimeout(reconnectTimer);
			reconnectTimer = undefined;
		}
		reconnectAttempts = 0;
		void connect();
	});

	void connect();

	return () => {
		closed = true;
		unsubscribeRecovered();
		if (reconnectTimer !== undefined) {
			clearTimeout(reconnectTimer);
		}
		ws?.close();
	};
}
