import {
	RealtimePresenceUpdatedSchema,
	type RealtimePresenceUpdated,
} from "@wichat/contracts/user/presence";
import { ensureAccessToken, getAccessToken } from "@/integrations/api-client";
import {
	notifyGatewayUnreachable,
	onGatewayRecovered,
} from "@/integrations/realtime/gateway-connection";
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
	let connecting = false;
	let refreshingAuth = false;
	let activeGroupId = options.groupId;
	let activity: "active" | "idle" = "active";
	let lastActivityMark = 0;
	let lastHeartbeatSentAt = 0;
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
		if (!ws || ws.readyState !== WebSocket.OPEN) {
			return;
		}
		lastHeartbeatSentAt = Date.now();
		const accessToken = getAccessToken();
		sendFrame({
			type: "presence.heartbeat",
			activity,
			...(accessToken ? { access_token: accessToken } : {}),
		});
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

	const resumePresence = () => {
		handleUserActivity();
		if (ws && ws.readyState === WebSocket.OPEN) {
			if (Date.now() - lastHeartbeatSentAt >= 10_000) {
				sendHeartbeat();
			}
			return;
		}
		if (!closed && !connecting && (!ws || ws.readyState === WebSocket.CLOSED)) {
			if (reconnectTimer) {
				clearTimeout(reconnectTimer);
				reconnectTimer = undefined;
			}
			reconnectAttempts = 0;
			void connect();
		}
	};

	const onVisibilityChange = () => {
		if (document.visibilityState === "visible") {
			resumePresence();
		}
	};

	const activityEvents = ["mousemove", "mousedown", "keydown", "touchstart", "scroll"] as const;
	if (typeof window !== "undefined") {
		for (const evt of activityEvents) {
			window.addEventListener(evt, handleUserActivity, { passive: true });
		}
		window.addEventListener("focus", resumePresence);
		window.addEventListener("online", resumePresence);
		document.addEventListener("visibilitychange", onVisibilityChange);
		scheduleIdleTimer();
	}

	const connect = async () => {
		if (closed || connecting) {
			return;
		}
		connecting = true;
		let token = getAccessToken();
		if (!token) {
			token = await ensureAccessToken();
		}
		if (closed) {
			connecting = false;
			return;
		}
		if (!token) {
			connecting = false;
			const delay = Math.min(1000 * 2 ** reconnectAttempts, 10_000);
			reconnectAttempts++;
			reconnectTimer = setTimeout(connect, delay);
			return;
		}

		try {
			ws = new WebSocket(realtimeWebSocketURL(token));
		} catch {
			connecting = false;
			const delay = Math.min(1000 * 2 ** reconnectAttempts, 10_000);
			reconnectAttempts++;
			reconnectTimer = setTimeout(connect, delay);
			return;
		}
		connecting = false;

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
				const raw = JSON.parse(String(event.data)) as { type?: string };
				if (raw.type === "presence.auth_required") {
					if (!refreshingAuth) {
						refreshingAuth = true;
						void ensureAccessToken()
							.then((nextToken) => {
								if (!closed && nextToken) {
									sendHeartbeat();
								}
							})
							.finally(() => {
								refreshingAuth = false;
							});
					}
					return;
				}
				const parsed = RealtimePresenceUpdatedSchema.parse(raw);
				if (parsed.type === "presence.updated") {
					options.onPresenceUpdated(parsed);
				}
			} catch {
				// ignore non-presence frames
			}
		});

		ws.addEventListener("close", () => {
			ws = null;
			if (heartbeatTimer) {
				clearInterval(heartbeatTimer);
				heartbeatTimer = undefined;
			}
			if (!closed) {
				notifyGatewayUnreachable();
				const delay = Math.min(1000 * 2 ** reconnectAttempts, 10_000);
				reconnectAttempts++;
				reconnectTimer = setTimeout(() => {
					reconnectTimer = undefined;
					void connect();
				}, delay);
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
		if (reconnectTimer) {
			clearTimeout(reconnectTimer);
			reconnectTimer = undefined;
		}
		reconnectAttempts = 0;
		void connect();
	});

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
			unsubscribeRecovered();
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
				window.removeEventListener("focus", resumePresence);
				window.removeEventListener("online", resumePresence);
				document.removeEventListener("visibilitychange", onVisibilityChange);
			}
			ws?.close();
			ws = null;
		},
	};
}
