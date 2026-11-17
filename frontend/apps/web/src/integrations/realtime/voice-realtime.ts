import {
	VoiceIdleWarningEventSchema,
	VoiceModerationAppliedEventSchema,
	VoiceOccupancyUpdatedEventSchema,
	VoiceRecordingUpdatedEventSchema,
	type VoiceIdleWarningEvent,
	type VoiceModerationAppliedEvent,
	type VoiceOccupancyUpdatedEvent,
	type VoiceRecordingUpdatedEvent,
} from "@wichat/contracts/chat/voice";

export type VoiceRealtimeEvent =
	| VoiceOccupancyUpdatedEvent
	| VoiceRecordingUpdatedEvent
	| VoiceModerationAppliedEvent
	| VoiceIdleWarningEvent;

type VoiceRealtimeListener = (event: VoiceRealtimeEvent) => void;

const listeners = new Set<VoiceRealtimeListener>();

export function subscribeVoiceRealtime(listener: VoiceRealtimeListener): () => void {
	listeners.add(listener);
	return () => listeners.delete(listener);
}

export function dispatchVoiceRealtimeFrame(raw: unknown): boolean {
	if (!raw || typeof raw !== "object" || !("type" in raw)) {
		return false;
	}
	const type = (raw as { type?: unknown }).type;
	if (typeof type !== "string" || !type.startsWith("voice.")) {
		return false;
	}

	let parsed: VoiceRealtimeEvent | null = null;
	try {
		if (type === "voice.occupancy.updated") {
			parsed = VoiceOccupancyUpdatedEventSchema.parse(raw);
		} else if (type === "voice.recording.updated") {
			parsed = VoiceRecordingUpdatedEventSchema.parse(raw);
		} else if (type === "voice.moderation.applied") {
			parsed = VoiceModerationAppliedEventSchema.parse(raw);
		} else if (type === "voice.idle.warning") {
			parsed = VoiceIdleWarningEventSchema.parse(raw);
		}
	} catch {
		return true;
	}

	if (parsed) {
		listeners.forEach((listener) => listener(parsed));
	}
	return true;
}
