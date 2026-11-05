import { isPublicEntityId } from "@wichat/contracts";

const STORAGE_KEY = "wichat.last_channel";

export type LastChannelRef = {
	groupId: string;
	channelId: string;
};

export function readLastChannel(): LastChannelRef | null {
	if (typeof window === "undefined") {
		return null;
	}
	try {
		const raw = window.localStorage.getItem(STORAGE_KEY);
		if (!raw) {
			return null;
		}
		const parsed = JSON.parse(raw) as LastChannelRef;
		if (isPublicEntityId(parsed?.groupId ?? "") && isPublicEntityId(parsed?.channelId ?? "")) {
			return parsed;
		}
	} catch {
		// ignore corrupt storage
	}
	return null;
}

export function writeLastChannel(ref: LastChannelRef) {
	if (typeof window === "undefined") {
		return;
	}
	window.localStorage.setItem(STORAGE_KEY, JSON.stringify(ref));
}

export function clearLastChannelIfGroup(groupId: string) {
	if (typeof window === "undefined") {
		return;
	}
	const last = readLastChannel();
	if (last?.groupId === groupId) {
		window.localStorage.removeItem(STORAGE_KEY);
	}
}
