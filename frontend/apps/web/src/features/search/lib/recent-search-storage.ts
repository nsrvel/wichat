const STORAGE_KEY = "wichat.recent_searches";
const MAX_ENTRIES = 12;

export function readRecentSearches(): string[] {
	if (typeof window === "undefined") {
		return [];
	}
	try {
		const raw = window.localStorage.getItem(STORAGE_KEY);
		if (!raw) {
			return [];
		}
		const parsed = JSON.parse(raw) as unknown;
		if (!Array.isArray(parsed)) {
			return [];
		}
		return parsed
			.filter((v): v is string => typeof v === "string")
			.map((s) => s.trim())
			.filter(Boolean)
			.slice(0, MAX_ENTRIES);
	} catch {
		return [];
	}
}

export function pushRecentSearch(query: string): string[] {
	const trimmed = query.trim();
	if (!trimmed || typeof window === "undefined") {
		return readRecentSearches();
	}
	const prev = readRecentSearches().filter((s) => s.toLowerCase() !== trimmed.toLowerCase());
	const next = [trimmed, ...prev].slice(0, MAX_ENTRIES);
	window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
	return next;
}

export function clearRecentSearches(): void {
	if (typeof window === "undefined") {
		return;
	}
	window.localStorage.removeItem(STORAGE_KEY);
}
