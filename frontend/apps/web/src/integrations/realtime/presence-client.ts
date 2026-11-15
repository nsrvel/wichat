import { RealtimePresenceUpdatedSchema, type RealtimePresenceUpdated } from "@wichat/contracts/user/presence";
import { ensureAccessToken, getAccessToken } from "@/integrations/api-client";
import { env } from "@/lib/env";

const HEARTBEAT_INTERVAL_MS = 25_000;
const IDLE_TIMEOUT_MS = 5 * 60 * 1000;

export type PresenceConnectionOptions = {
	groupId?: string;
	onPresenceUpdated: (event: RealtimePresenceUpdated) => void;
};

function realtimeWebSocketURL(accessToken: string): string {
	const base = env.NEXT_PUBLIC_API_URL.replace(/\/+$/, "").replace(/^http/, "ws");
	const pathname = base.endsWith("/api/v1") ? `${base}/realtime` : `${base}/api/v1/realtime`;
	const url = new URL(pathname);
	url.searchParams.set("access_token", accessToken);
	return url.toString();
}

/**
 * Maintains a persistent workspace WebSocket for presence heartbeats, automatic 5-minute idle
 * detection, and live `presence.updated` fanout for the active group.
 */
export function connectPresenceRealtime(options: PresenceConnectionOptions): {
	setGroupId: (groupId?: string) => void;
	close: () => void;
} {
	let ws: WebSocket | null = null;
	let closed = false;
	let activeGroupId = options.groupId;
	let activity: "active" | "idle" = "active";
	let lastActivityMark = 0;
	let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
	let heartbeatTimer: ReturnType<typeof setInterval> | undefined;
	let idleTimer: ReturnType<typeof setTimeout> | undefined;
	let reconnectAttempts = 0;

	const sendFrame = (payload: Record<string, unknown>) => {
		if (ws && ws.readyState === WebSocket.OPEN) {
			ws.send(JSON.stringify(payload));
		}
	};

	const sendHeartbeat = () => {
		sendFrame({ type: "presence.heartbeat", activity });
	};

	const sendGroupSubscribe = () => {
		if (activeGroupId) {
			sendFrame({ type: "subscribe_group", group_id: activeGroupId });
		}
	};

	const scheduleIdleTimer = () => {
		if (idleTimer) {
			clearTimeout(idleTimer);
		}
		idleTimer = setTimeout(() => {
			if (activity !== "idle") {
				activity = "idle";
				sendHeartbeat();
			}
		}, IDLE_TIMEOUT_MS);
	};

	const handleUserActivity = () => {
		const now = Date.now();
		if (activity === "idle") {
			activity = "active";
			lastActivityMark = now;
			sendHeartbeat();
			scheduleIdleTimer();
			return;
		}
		if (now - lastActivityMark >= 1_000) {
			lastActivityMark = now;
			scheduleIdleTimer();
		}
	};

	const onVisibilityChange = () => {
		if (document.visibilityState === "visible") {
			handleUserActivity();
		}
	};

	const activityEvents = ["mousemove", "mousedown", "keydown", "touchstart", "scroll", "focus"] as const;
	if (typeof window !== "undefined") {
		for (const evt of activityEvents) {
			window.addEventListener(evt, handleUserActivity, { passive: true });
		}
		document.addEventListener("visibilitychange", onVisibilityChange);
		scheduleIdleTimer();
	}

	const connect = async () => {
		if (closed) {
			return;
		}
		let token = getAccessToken();
		if (!token) {
			token = await ensureAccessToken();
		}
		if (!token) {
			const delay = Math.min(1000 * 2 ** reconnectAttempts, 10_000);
			reconnectAttempts++;
			reconnectTimer = setTimeout(connect, delay);
			return;
		}

		try {
			ws = new WebSocket(realtimeWebSocketURL(token));
		} catch {
			const delay = Math.min(1000 * 2 ** reconnectAttempts, 10_000);
			reconnectAttempts++;
			reconnectTimer = setTimeout(connect, delay);
			return;
		}

		ws.addEventListener("open", () => {
			reconnectAttempts = 0;
			sendGroupSubscribe();
			sendHeartbeat();
			if (heartbeatTimer) {
				clearInterval(heartbeatTimer);
			}
			heartbeatTimer = setInterval(sendHeartbeat, HEARTBEAT_INTERVAL_MS);
		});

		ws.addEventListener("message", (event) => {
			try {
				const parsed = RealtimePresenceUpdatedSchema.parse(JSON.parse(String(event.data)));
				if (parsed.type === "presence.updated") {
					options.onPresenceUpdated(parsed);
				}
			} catch {
				// ignore non-presence frames
			}
		});

		ws.addEventListener("close", () => {
			if (heartbeatTimer) {
				clearInterval(heartbeatTimer);
				heartbeatTimer = undefined;
			}
			if (!closed) {
				const delay = Math.min(1000 * 2 ** reconnectAttempts, 10_000);
				reconnectAttempts++;
				reconnectTimer = setTimeout(connect, delay);
			}
		});
	};

	void connect();

	return {
		setGroupId: (nextGroupId?: string) => {
			if (nextGroupId === activeGroupId) {
				return;
			}
			activeGroupId = nextGroupId;
			sendGroupSubscribe();
		},
		close: () => {
			closed = true;
			if (reconnectTimer) {
				clearTimeout(reconnectTimer);
			}
			if (heartbeatTimer) {
				clearInterval(heartbeatTimer);
			}
			if (idleTimer) {
				clearTimeout(idleTimer);
			}
			if (typeof window !== "undefined") {
				for (const evt of activityEvents) {
					window.removeEventListener(evt, handleUserActivity);
				}
				document.removeEventListener("visibilitychange", onVisibilityChange);
			}
			ws?.close();
			ws = null;
		},
	};
}
