import { RealtimeMessageCreatedSchema, type Message } from "@wichat/contracts";
import { ensureAccessToken, getAccessToken } from "@/integrations/api-client";
import { env } from "@/lib/env";

export type ChatChannelRealtimeOptions = {
	groupId: string;
	channelId: string;
	onMessageCreated: (message: Message) => void;
};

function realtimeWebSocketURL(accessToken: string): string {
	const base = env.NEXT_PUBLIC_API_URL.replace(/\/+$/, "").replace(/^http/, "ws");
	const pathname = base.endsWith("/api/v1") ? `${base}/realtime` : `${base}/api/v1/realtime`;
	const url = new URL(pathname);
	url.searchParams.set("access_token", accessToken);
	return url.toString();
}

/** Opens a WebSocket, subscribes to one channel, and forwards message.created events. */
export function subscribeChatChannel(options: ChatChannelRealtimeOptions): () => void {
	let ws: WebSocket | null = null;
	let closed = false;
	let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
	let reconnectAttempts = 0;

	const connect = async () => {
		if (closed) {
			return;
		}
		let token = getAccessToken();
		if (!token) {
			token = await ensureAccessToken();
		}
		if (!token) {
			const delay = Math.min(1000 * 2 ** reconnectAttempts, 10000);
			reconnectAttempts++;
			reconnectTimer = setTimeout(connect, delay);
			return;
		}

		try {
			ws = new WebSocket(realtimeWebSocketURL(token));
		} catch {
			const delay = Math.min(1000 * 2 ** reconnectAttempts, 10000);
			reconnectAttempts++;
			reconnectTimer = setTimeout(connect, delay);
			return;
		}

		ws.addEventListener("open", () => {
			reconnectAttempts = 0;
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
				const delay = Math.min(1000 * 2 ** reconnectAttempts, 10000);
				reconnectAttempts++;
				reconnectTimer = setTimeout(connect, delay);
			}
		});
	};

	void connect();

	return () => {
		closed = true;
		if (reconnectTimer) {
			clearTimeout(reconnectTimer);
		}
		ws?.close();
		ws = null;
	};
}
