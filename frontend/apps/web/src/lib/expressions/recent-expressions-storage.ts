import type { GifEntry, StickerEntry } from "@/lib/expressions/expression-types";

const RECENT_EMOJIS_KEY = "wichat:recent-emojis:v1";
const RECENT_GIFS_KEY = "wichat:recent-gifs:v1";
const RECENT_STICKERS_KEY = "wichat:recent-stickers:v1";
const SKIN_TONE_KEY = "wichat:emoji-skin-tone:v1";

const MAX_RECENT_EMOJIS = 24;
const MAX_RECENT_MEDIA = 12;

const EMPTY_EMOJIS: string[] = [];
const EMPTY_GIFS: GifEntry[] = [];
const EMPTY_STICKERS: StickerEntry[] = [];

const listeners = new Set<() => void>();

type SnapshotCache<T> = {
	raw: string | null;
	value: T;
};

const emojiCache: SnapshotCache<string[]> = { raw: null, value: EMPTY_EMOJIS };
const gifCache: SnapshotCache<GifEntry[]> = { raw: null, value: EMPTY_GIFS };
const stickerCache: SnapshotCache<StickerEntry[]> = { raw: null, value: EMPTY_STICKERS };

function notifyListeners() {
	for (const listener of listeners) {
		listener();
	}
}

export function subscribeRecentExpressions(listener: () => void): () => void {
	listeners.add(listener);
	return () => {
		listeners.delete(listener);
	};
}

function readCachedJson<T>(key: string, cache: SnapshotCache<T>, fallback: T): T {
	if (typeof window === "undefined") {
		return fallback;
	}
	try {
		const raw = window.localStorage.getItem(key);
		if (raw === cache.raw) {
			return cache.value;
		}
		if (!raw) {
			cache.raw = null;
			cache.value = fallback;
			return fallback;
		}
		const parsed = JSON.parse(raw) as T;
		cache.raw = raw;
		cache.value = parsed;
		return parsed;
	} catch {
		return cache.value;
	}
}

function writeCachedJson<T>(key: string, cache: SnapshotCache<T> | null, value: T) {
	if (typeof window === "undefined") {
		return;
	}
	try {
		const raw = JSON.stringify(value);
		window.localStorage.setItem(key, raw);
		if (cache) {
			cache.raw = raw;
			cache.value = value;
		}
		notifyListeners();
	} catch {
		// Ignore quota / private-mode storage errors.
	}
}

export function getRecentEmojis(): string[] {
	return readCachedJson(RECENT_EMOJIS_KEY, emojiCache, EMPTY_EMOJIS);
}

export function recordRecentEmoji(emoji: string) {
	const current = getRecentEmojis().filter((item) => item !== emoji);
	writeCachedJson(RECENT_EMOJIS_KEY, emojiCache, [emoji, ...current].slice(0, MAX_RECENT_EMOJIS));
}

export function getRecentGifs(): GifEntry[] {
	return readCachedJson(RECENT_GIFS_KEY, gifCache, EMPTY_GIFS);
}

export function recordRecentGif(gif: GifEntry) {
	const current = getRecentGifs().filter((item) => item.id !== gif.id);
	writeCachedJson(RECENT_GIFS_KEY, gifCache, [gif, ...current].slice(0, MAX_RECENT_MEDIA));
}

export function getRecentStickers(): StickerEntry[] {
	return readCachedJson(RECENT_STICKERS_KEY, stickerCache, EMPTY_STICKERS);
}

export function recordRecentSticker(sticker: StickerEntry) {
	const current = getRecentStickers().filter((item) => item.id !== sticker.id);
	writeCachedJson(
		RECENT_STICKERS_KEY,
		stickerCache,
		[sticker, ...current].slice(0, MAX_RECENT_MEDIA),
	);
}

export function getSavedSkinTone(): number {
	if (typeof window === "undefined") {
		return 0;
	}
	try {
		const raw = window.localStorage.getItem(SKIN_TONE_KEY);
		if (!raw) {
			return 0;
		}
		const parsed = Number(JSON.parse(raw));
		return Number.isFinite(parsed) ? parsed : 0;
	} catch {
		return 0;
	}
}

export function saveSkinTone(tone: number) {
	writeCachedJson(SKIN_TONE_KEY, null, tone);
}
