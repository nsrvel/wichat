import {
	RealtimeChatEventSchema,
	type Message,
	type MessageReactionSummary,
} from "@wichat/contracts/chat/message";
import { RealtimePresenceUpdatedSchema } from "@wichat/contracts/user/presence";
import {
	ensureAccessToken,
	forceRefreshAccessToken,
	getAccessToken,
} from "@/integrations/api-client";
import { onGatewayRecovered } from "@/integrations/realtime/gateway-connection";
import { normalizePublicGroupId } from "@/features/presence/public-group-id";
import { realtimeHttpUrl } from "@/lib/realtime-url";

/** Short marker only — Chromium rejects `bearer.<jwt>` when the JWT exceeds subprotocol length. */
export function realtimeWsSubprotocols(): string[] {
	return ["bearer"];
}

export function buildRealtimeWsUrl(): string {
	return realtimeHttpUrl().replace(/^http/, "ws");
}

const HEARTBEAT_MS = 25_000;
const IDLE_MS = 5 * 60_000;

export type LivePresenceListener = (frame: {
	user_id: string;
	presence_status: "online" | "idle" | "dnd" | "offline";
	is_online: boolean;
}) => void;

export type ChannelRealtimeHandlers = {
	groupId: string;
	channelId: string;
	onMessageCreated: (message: Message) => void;
	onMessageUpdated?: (message: Message) => void;
	onMessageDeleted?: (messageId: string) => void;
	onMessageReactionsUpdated?: (messageId: string, reactions: MessageReactionSummary[]) => void;
	onReconnect?: () => void;
};

type Runtime = {
	ws: WebSocket | null;
	closed: boolean;
	connecting: boolean;
	reconnectTimer?: ReturnType<typeof setTimeout>;
	heartbeatTimer?: ReturnType<typeof setInterval>;
	hasConnectedOnce: boolean;
	reconnectAttempts: number;
	groupId?: string;
	presenceListeners: Set<LivePresenceListener>;
	channel?: ChannelRealtimeHandlers;
	activity: "active" | "idle";
	refCount: number;
};

let rt: Runtime | null = null;
let presenceAuthRequiredStreak = 0;

async function handlePresenceAuthRequired(ws: WebSocket): Promise<void> {
	const token = (await forceRefreshAccessToken()) || (await ensureAccessToken());
	if (!token || ws.readyState !== WebSocket.OPEN) {
		return;
	}
	sendJson(ws, {
		type: "presence.heartbeat",
		activity: rt?.activity ?? "active",
		access_token: token,
	});
}

function ensureRuntime(): Runtime {
	if (!rt) {
		rt = {
			ws: null,
			closed: false,
			connecting: false,
			hasConnectedOnce: false,
			reconnectAttempts: 0,
			presenceListeners: new Set(),
			activity: "active",
			refCount: 0,
		};
	}
	return rt;
}

function sendJson(ws: WebSocket, payload: Record<string, unknown>): void {
	if (ws.readyState === WebSocket.OPEN) {
		ws.send(JSON.stringify(payload));
	}
}

async function sendPresenceHeartbeat(ws: WebSocket): Promise<void> {
	if (ws.readyState !== WebSocket.OPEN) {
		return;
	}
	const token = (await ensureAccessToken()) || getAccessToken();
	sendJson(ws, {
		type: "presence.heartbeat",
		activity: rt?.activity ?? "active",
		...(token ? { access_token: token } : {}),
	});
}

function heartbeat(ws: WebSocket): void {
	void sendPresenceHeartbeat(ws);
}

function subscribePresence(ws: WebSocket): void {
	const groupId = normalizePublicGroupId(rt?.groupId);
	if (groupId) {
		sendJson(ws, { type: "subscribe_group", group_id: groupId });
	}
}

function subscribeChannel(ws: WebSocket): void {
	const ch = rt?.channel;
	if (!ch) {
		return;
	}
	sendJson(ws, {
		type: "subscribe",
		group_id: ch.groupId,
		channel_id: ch.channelId,
	});
}

function dispatchChatFrame(raw: unknown): void {
	const ch = rt?.channel;
	if (!ch) {
		return;
	}
	try {
		const parsed = RealtimeChatEventSchema.parse(raw);
		if (parsed.type === "message.created") {
			ch.onMessageCreated(parsed.message);
		} else if (parsed.type === "message.updated") {
			ch.onMessageUpdated?.(parsed.message);
		} else if (parsed.type === "message.deleted") {
			ch.onMessageDeleted?.(parsed.message_id);
		} else if (parsed.type === "message.reactions.updated") {
			ch.onMessageReactionsUpdated?.(parsed.message_id, parsed.reactions);
		}
	} catch {
		// not a chat frame
	}
}

function scheduleReconnect(): void {
	const state = ensureRuntime();
	if (state.closed || state.reconnectTimer) {
		return;
	}
	const delay = Math.min(1000 * 2 ** state.reconnectAttempts, 10_000);
	state.reconnectAttempts++;
	state.reconnectTimer = setTimeout(() => {
		state.reconnectTimer = undefined;
		void connect();
	}, delay);
}

async function connect(): Promise<void> {
	const state = ensureRuntime();
	if (state.closed || state.connecting || state.ws?.readyState === WebSocket.OPEN) {
		return;
	}
	if (state.refCount === 0) {
		return;
	}

	state.connecting = true;
	await ensureAccessToken().catch(() => null);
	state.connecting = false;

	if (state.closed) {
		return;
	}

	const wsUrl = buildRealtimeWsUrl();
	let ws: WebSocket;
	try {
		ws = new WebSocket(wsUrl, realtimeWsSubprotocols());
	} catch (err) {
		if (process.env.NODE_ENV === "development") {
			console.warn("[wichat] websocket constructor failed", wsUrl, err);
		}
		scheduleReconnect();
		return;
	}
	state.ws = ws;

	ws.addEventListener("open", () => {
		state.reconnectAttempts = 0;
		presenceAuthRequiredStreak = 0;
		void (async () => {
			const token = (await forceRefreshAccessToken()) || (await ensureAccessToken());
			if (!token || ws.readyState !== WebSocket.OPEN) {
				ws.close(4003, "missing access token");
				scheduleReconnect();
				return;
			}
			await sendPresenceHeartbeat(ws);
			subscribePresence(ws);
			subscribeChannel(ws);
			heartbeat(ws);
			if (state.heartbeatTimer) {
				clearInterval(state.heartbeatTimer);
			}
			state.heartbeatTimer = setInterval(() => heartbeat(ws), HEARTBEAT_MS);
			if (state.hasConnectedOnce) {
				state.channel?.onReconnect?.();
			}
			state.hasConnectedOnce = true;
		})();
	});

	ws.addEventListener("error", () => {
		if (process.env.NODE_ENV === "development") {
			console.warn("[wichat] websocket error", wsUrl);
		}
	});

	ws.addEventListener("message", (event) => {
		let raw: unknown;
		try {
			raw = JSON.parse(String(event.data));
		} catch {
			return;
		}
		const head = raw as { type?: string };
		if (
			head.type === "message.created" ||
			head.type === "message.updated" ||
			head.type === "message.deleted" ||
			head.type === "message.reactions.updated"
		) {
			dispatchChatFrame(raw);
			return;
		}
		if (head.type === "subscribe.denied") {
			return;
		}
		if (head.type === "presence.auth_required") {
			presenceAuthRequiredStreak += 1;
			void handlePresenceAuthRequired(ws);
			if (presenceAuthRequiredStreak >= 3) {
				presenceAuthRequiredStreak = 0;
				ws.close(4001, "presence auth refresh");
			}
			return;
		}
		try {
			const parsed = RealtimePresenceUpdatedSchema.parse(raw);
			presenceAuthRequiredStreak = 0;
			const frame = {
				user_id: parsed.user_id,
				presence_status: parsed.presence_status,
				is_online: parsed.is_online,
			};
			for (const listener of state.presenceListeners) {
				listener(frame);
			}
		} catch {
			// ignore
		}
	});

	ws.addEventListener("close", (event) => {
		if (process.env.NODE_ENV === "development" && event.code !== 1000) {
			console.warn("[wichat] websocket closed", wsUrl, event.code, event.reason);
		}
		state.ws = null;
		if (state.heartbeatTimer) {
			clearInterval(state.heartbeatTimer);
			state.heartbeatTimer = undefined;
		}
		if (!state.closed) {
			scheduleReconnect();
		}
	});
}

function acquire(): void {
	const state = ensureRuntime();
	state.refCount++;
	state.closed = false;
	if (state.refCount === 1) {
		if (!unbindActivity) {
			unbindActivity = bindActivity();
		}
		void connect();
		const unsub = onGatewayRecovered(() => {
			if (state.closed || state.refCount === 0) {
				return;
			}
			if (
				state.ws?.readyState === WebSocket.OPEN ||
				state.ws?.readyState === WebSocket.CONNECTING
			) {
				return;
			}
			if (state.reconnectTimer) {
				clearTimeout(state.reconnectTimer);
				state.reconnectTimer = undefined;
			}
			state.reconnectAttempts = 0;
			void connect();
		});
		stateAcquireCleanup = unsub;
	}
}

let stateAcquireCleanup: (() => void) | undefined;
let unbindActivity: (() => void) | undefined;

function bindActivity(): () => void {
	const events = ["mousemove", "mousedown", "keydown", "touchstart", "scroll"] as const;
	let lastResetAt = Date.now();
	let idleTimer = setTimeout(() => setGatewaySocketActivity("idle"), IDLE_MS);
	const resetIdle = () => {
		const now = Date.now();
		if (rt?.activity !== "active") {
			setGatewaySocketActivity("active");
		} else if (now - lastResetAt < 5_000) {
			return;
		}
		lastResetAt = now;
		clearTimeout(idleTimer);
		idleTimer = setTimeout(() => setGatewaySocketActivity("idle"), IDLE_MS);
	};
	for (const e of events) {
		window.addEventListener(e, resetIdle, { passive: true });
	}
	window.addEventListener("focus", resetIdle);
	return () => {
		for (const e of events) {
			window.removeEventListener(e, resetIdle);
		}
		window.removeEventListener("focus", resetIdle);
		clearTimeout(idleTimer);
	};
}

function release(): void {
	const state = ensureRuntime();
	state.refCount = Math.max(0, state.refCount - 1);
	if (state.refCount === 0) {
		state.closed = true;
		unbindActivity?.();
		unbindActivity = undefined;
		stateAcquireCleanup?.();
		stateAcquireCleanup = undefined;
		if (state.reconnectTimer) {
			clearTimeout(state.reconnectTimer);
		}
		if (state.heartbeatTimer) {
			clearInterval(state.heartbeatTimer);
		}
		state.ws?.close();
		state.ws = null;
		state.channel = undefined;
		rt = null;
	}
}

export function acquireGatewaySocketPresence(listener: LivePresenceListener): () => void {
	acquire();
	const state = ensureRuntime();
	state.presenceListeners.add(listener);
	return () => {
		state.presenceListeners.delete(listener);
		release();
	};
}

export function getGatewaySocketGroupId(): string | undefined {
	return rt?.groupId;
}

export function setGatewaySocketGroupId(groupId?: string): void {
	const state = ensureRuntime();
	const next = normalizePublicGroupId(groupId);
	if (state.groupId === next) {
		return;
	}
	state.groupId = next;
	if (state.ws?.readyState === WebSocket.OPEN) {
		subscribePresence(state.ws);
	}
}

export function acquireGatewaySocketChannel(handlers: ChannelRealtimeHandlers): () => void {
	acquire();
	const state = ensureRuntime();
	state.channel = handlers;
	if (state.ws?.readyState === WebSocket.OPEN) {
		subscribeChannel(state.ws);
	}
	return () => {
		if (state.channel === handlers) {
			state.channel = undefined;
		}
		release();
	};
}

/** After manual presence PUT, push a heartbeat so ms-user registers the open conn. */
export function nudgePresenceHeartbeat(): void {
	const state = rt;
	if (!state?.ws || state.ws.readyState !== WebSocket.OPEN) {
		return;
	}
	void ensureAccessToken().then((token) => {
		if (!token || state.ws?.readyState !== WebSocket.OPEN) {
			return;
		}
		heartbeat(state.ws);
	});
}

export function setGatewaySocketActivity(activity: "active" | "idle"): void {
	const state = ensureRuntime();
	if (state.activity === activity) {
		return;
	}
	state.activity = activity;
	if (state.ws?.readyState === WebSocket.OPEN) {
		heartbeat(state.ws);
	}
}
