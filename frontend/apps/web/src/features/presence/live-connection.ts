import {
	acquireGatewaySocketPresence,
	buildRealtimeWsUrl,
	setGatewaySocketGroupId,
} from "@/integrations/realtime/gateway-socket";

export type LivePresenceListener = (frame: {
	user_id: string;
	presence_status: "online" | "idle" | "dnd" | "offline";
	is_online: boolean;
}) => void;

export { buildRealtimeWsUrl };

export function acquireLivePresence(listener: LivePresenceListener): () => void {
	return acquireGatewaySocketPresence(listener);
}

export function setLivePresenceGroupId(groupId?: string): void {
	setGatewaySocketGroupId(groupId);
}
