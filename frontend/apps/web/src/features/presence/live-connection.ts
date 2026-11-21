import { RealtimePresenceUpdatedSchema } from "@wichat/contracts/user/presence";
import { ensureAccessToken, getAccessToken } from "@/integrations/api-client";
import { normalizePublicGroupId } from "@/features/presence/public-group-id";
import { env } from "@/lib/env";

const HEARTBEAT_MS = 25_000;
const IDLE_MS = 5 * 60_000;

export type LivePresenceListener = (frame: {
	user_id: string;
	presence_status: "online" | "idle" | "dnd" | "offline";
	is_online: boolean;
}) => void;

function realtimeHttpUrl(): string {
	const base = env.NEXT_PUBLIC_API_URL.replace(/\/+$/, "");
	return base.endsWith("/api/v1") ? `${base}/realtime` : `${base}/api/v1/realtime`;
}

/** Access token on query string — avoids brittle `Sec-WebSocket-Protocol` + JWT size limits. */
export function buildRealtimeWsUrl(token: string): string {
	const url = new URL(realtimeHttpUrl());
	url.searchParams.set("access_token", token);
	return url.toString().replace(/^http/, "ws");
}

type Runtime = {
	ws: WebSocket | null;
	closed: boolean;
	groupId?: string;
	listeners: Set<LivePresenceListener>;
	reconnectTimer?: ReturnType<typeof setTimeout>;
	heartbeatTimer?: ReturnType<typeof setInterval>;
	idleTimer?: ReturnType<typeof setTimeout>;
	activity: "active" | "idle";
	connecting: boolean;
};

let rt: Runtime | null = null;

function ensureRuntime(): Runtime {
	if (!rt) {
		rt = {
			ws: null,
			closed: false,
			listeners: new Set(),
			activity: "active",
			connecting: false,
		};
	}
	return rt;
}

function sendJson(ws: WebSocket, payload: Record<string, unknown>): void {
	if (ws.readyState === WebSocket.OPEN) {
		ws.send(JSON.stringify(payload));
	}
}

function heartbeat(ws: WebSocket): void {
	const token = getAccessToken();
	sendJson(ws, {
		type: "presence.heartbeat",
		activity: rt?.activity ?? "active",
		...(token ? { access_token: token } : {}),
	});
}

function subscribeGroup(ws: WebSocket): void {
	const groupId = normalizePublicGroupId(rt?.groupId);
	if (groupId) {
		sendJson(ws, { type: "subscribe_group", group_id: groupId });
	}
}

function scheduleIdle(): void {
	if (!rt) {
		return;
	}
	if (rt.idleTimer) {
		clearTimeout(rt.idleTimer);
	}
	rt.idleTimer = setTimeout(() => {
		if (!rt || rt.activity === "idle") {
			return;
		}
		rt.activity = "idle";
		if (rt.ws) {
			heartbeat(rt.ws);
		}
	}, IDLE_MS);
}

function bindActivity(): () => void {
	const onActive = () => {
		if (!rt) {
			return;
		}
		if (rt.activity === "idle") {
			rt.activity = "active";
			if (rt.ws) {
				heartbeat(rt.ws);
			}
		}
		scheduleIdle();
	};
	const events = ["mousemove", "mousedown", "keydown", "touchstart", "scroll"] as const;
	for (const e of events) {
		window.addEventListener(e, onActive, { passive: true });
	}
	window.addEventListener("focus", onActive);
	document.addEventListener("visibilitychange", () => {
		if (document.visibilityState === "visible") {
			onActive();
		}
	});
	scheduleIdle();
	return () => {
		for (const e of events) {
			window.removeEventListener(e, onActive);
		}
		window.removeEventListener("focus", onActive);
	};
}

let unbindActivity: (() => void) | undefined;

async function connectLoop(): Promise<void> {
	const state = ensureRuntime();
	if (state.closed || state.connecting || state.ws?.readyState === WebSocket.OPEN) {
		return;
	}
	state.connecting = true;
	const token = getAccessToken() ?? (await ensureAccessToken());
	state.connecting = false;
	if (state.closed || !token) {
		state.reconnectTimer = setTimeout(() => void connectLoop(), 3_000);
		return;
	}

	let ws: WebSocket;
	try {
		ws = new WebSocket(buildRealtimeWsUrl(token));
	} catch {
		state.reconnectTimer = setTimeout(() => void connectLoop(), 3_000);
		return;
	}
	state.ws = ws;

	ws.addEventListener("open", () => {
		subscribeGroup(ws);
		heartbeat(ws);
		if (state.heartbeatTimer) {
			clearInterval(state.heartbeatTimer);
		}
		state.heartbeatTimer = setInterval(() => heartbeat(ws), HEARTBEAT_MS);
	});

	ws.addEventListener("message", (event) => {
		try {
			const raw = JSON.parse(String(event.data)) as { type?: string };
			if (raw.type === "subscribe.denied" || raw.type === "presence.auth_required") {
				return;
			}
			const parsed = RealtimePresenceUpdatedSchema.parse(raw);
			const frame = {
				user_id: parsed.user_id,
				presence_status: parsed.presence_status,
				is_online: parsed.is_online,
			};
			for (const listener of state.listeners) {
				listener(frame);
			}
		} catch {
			// ignore
		}
	});

	ws.addEventListener("close", () => {
		state.ws = null;
		if (state.heartbeatTimer) {
			clearInterval(state.heartbeatTimer);
			state.heartbeatTimer = undefined;
		}
		if (!state.closed) {
			state.reconnectTimer = setTimeout(() => void connectLoop(), 3_000);
		}
	});
}

export function acquireLivePresence(listener: LivePresenceListener): () => void {
	const state = ensureRuntime();
	state.listeners.add(listener);
	if (state.listeners.size === 1) {
		unbindActivity = bindActivity();
		void connectLoop();
	}
	return () => {
		state.listeners.delete(listener);
		if (state.listeners.size === 0) {
			state.closed = true;
			unbindActivity?.();
			unbindActivity = undefined;
			if (state.reconnectTimer) {
				clearTimeout(state.reconnectTimer);
			}
			if (state.heartbeatTimer) {
				clearInterval(state.heartbeatTimer);
			}
			if (state.idleTimer) {
				clearTimeout(state.idleTimer);
			}
			state.ws?.close();
			rt = null;
		}
	};
}

export function setLivePresenceGroupId(groupId?: string): void {
	const state = ensureRuntime();
	const next = normalizePublicGroupId(groupId);
	if (state.groupId === next) {
		return;
	}
	state.groupId = next;
	if (state.ws && state.ws.readyState === WebSocket.OPEN) {
		subscribeGroup(state.ws);
	}
}
