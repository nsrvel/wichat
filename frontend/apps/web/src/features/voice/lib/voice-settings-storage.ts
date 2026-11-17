import type { VoiceUserSettings } from "@wichat/contracts/chat/voice";

const STORAGE_KEY = "wichat.voice.settings.v1";

export const DEFAULT_VOICE_SETTINGS: VoiceUserSettings = {
	inputDeviceId: "",
	outputDeviceId: "",
	cameraDeviceId: "",
	inputMode: "vad",
	vadThreshold: 40,
	pttKey: "Space",
	echoCancellation: true,
	noiseSuppression: true,
	autoGainControl: true,
	userVolumes: {},
	locallyMutedUsers: {},
};

const listeners = new Set<() => void>();
let cachedRaw: string | null = null;
let cachedParsed: VoiceUserSettings = DEFAULT_VOICE_SETTINGS;

export function readVoiceSettings(): VoiceUserSettings {
	if (typeof window === "undefined") return DEFAULT_VOICE_SETTINGS;
	const raw = window.localStorage.getItem(STORAGE_KEY);
	if (raw === cachedRaw) return cachedParsed;
	cachedRaw = raw;
	if (!raw) {
		cachedParsed = DEFAULT_VOICE_SETTINGS;
		return cachedParsed;
	}
	try {
		const parsed = JSON.parse(raw) as Partial<VoiceUserSettings>;
		cachedParsed = {
			...DEFAULT_VOICE_SETTINGS,
			...parsed,
			userVolumes: { ...(parsed.userVolumes ?? {}) },
			locallyMutedUsers: { ...(parsed.locallyMutedUsers ?? {}) },
		};
	} catch {
		cachedParsed = DEFAULT_VOICE_SETTINGS;
	}
	return cachedParsed;
}

export function writeVoiceSettings(patch: Partial<VoiceUserSettings>) {
	if (typeof window === "undefined") return;
	const next: VoiceUserSettings = {
		...readVoiceSettings(),
		...patch,
	};
	const raw = JSON.stringify(next);
	cachedRaw = raw;
	cachedParsed = next;
	window.localStorage.setItem(STORAGE_KEY, raw);
	listeners.forEach((fn) => fn());
}

export function subscribeVoiceSettings(listener: () => void): () => void {
	listeners.add(listener);
	return () => listeners.delete(listener);
}
