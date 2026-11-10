import { RealtimeMessageCreatedSchema, type Message } from "@wichat/contracts";
import { getAccessToken } from "@/integrations/api-client";
import { env } from "@/env";

export type ChatChannelRealtimeOptions = {
	groupId: string;
	channelId: string;
	onMessageCreated: (message: Message) => void;
};

function realtimeWebSocketURL(accessToken: string): string {
	const base = env.NEXT_PUBLIC_API_URL.replace(/^http/, "ws");
	const url = new URL(`${base}/realtime`);
	url.searchParams.set("access_token", accessToken);
	return url.toString();
}

/** Opens a WebSocket, subscribes to one channel, and forwards message.created events. */
export function subscribeChatChannel(options: ChatChannelRealtimeOptions): () => void {
	let ws: WebSocket | null = null;
	let closed = false;
	let reconnectTimer: ReturnType<typeof setTimeout> | undefined;

	const connect = () => {
		if (closed) {
			return;
		}
		const token = getAccessToken();
		if (!token) {
			reconnectTimer = setTimeout(connect, 2000);
			return;
		}

		ws = new WebSocket(realtimeWebSocketURL(token));
		ws.addEventListener("open", () => {
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
				const parsed = RealtimeMessageCreatedSchema.parse(JSON.parse(String(event.data)));
				if (parsed.type === "message.created") {
					options.onMessageCreated(parsed.message);
				}
			} catch {
				// ignore malformed fanout frames
			}
		});
		ws.addEventListener("close", () => {
			if (!closed) {
				reconnectTimer = setTimeout(connect, 2000);
			}
		});
	};

	connect();

	return () => {
		closed = true;
		if (reconnectTimer) {
			clearTimeout(reconnectTimer);
		}
		ws?.close();
		ws = null;
	};
}
