import {
	acquireGatewaySocketChannels,
	acquireGatewaySocketPresence,
	buildRealtimeWsUrl,
	getGatewaySocketGroupId,
	setGatewaySocketGroupId,
	type LiveChannelsListener,
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

export function acquireLiveChannels(listener: LiveChannelsListener): () => void {
	return acquireGatewaySocketChannels(listener);
}

export function getLivePresenceGroupId(): string | undefined {
	return getGatewaySocketGroupId();
}

export function setLivePresenceGroupId(groupId?: string): void {
	setGatewaySocketGroupId(groupId);
}
