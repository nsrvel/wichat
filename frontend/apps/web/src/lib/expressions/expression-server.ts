import type {
	MediaExpressionCategory,
	MediaExpressionEntry,
	MediaExpressionKind,
} from "@/lib/expressions/expression-types";

const DEFAULT_MEDIA_CATEGORIES: Record<
	MediaExpressionKind,
	Array<{ id: string; label: string; query: string }>
> = {
	gifs: [
		{ id: "trending", label: "Trending", query: "" },
		{ id: "reactions", label: "Reactions", query: "reaction" },
		{ id: "greetings", label: "Greetings", query: "hello" },
		{ id: "perfect", label: "Perfect", query: "perfect" },
		{ id: "welcome", label: "Welcome", query: "welcome" },
		{ id: "congratulations", label: "Congratulations", query: "congratulations" },
		{ id: "thank-you", label: "Thank you", query: "thank you" },
		{ id: "excited", label: "Excited", query: "excited" },
	],
	stickers: [
		{ id: "trending", label: "Popular", query: "" },
		{ id: "greetings", label: "Greetings", query: "hello" },
		{ id: "reactions", label: "Reactions", query: "reaction" },
		{ id: "celebrate", label: "Celebrate", query: "party" },
		{ id: "work", label: "Work Mood", query: "coffee" },
		{ id: "love", label: "Love", query: "love" },
		{ id: "funny", label: "Funny", query: "funny" },
		{ id: "thanks", label: "Thank you", query: "thank you" },
	],
};

export type ExpressionMediaPayload = {
	configured: boolean;
	ok: boolean;
	provider: "klipy" | "giphy" | "none";
	items: MediaExpressionEntry[];
	categories: MediaExpressionCategory[];
};

const LIST_CACHE_TTL_MS = 3 * 60 * 1000;
const CATEGORIES_CACHE_TTL_MS = 15 * 60 * 1000;
const MAX_CACHE_ENTRIES = 120;

const responseCache = new Map<string, { expiresAt: number; payload: ExpressionMediaPayload }>();
const inflightRequests = new Map<string, Promise<ExpressionMediaPayload>>();

function getCachedPayload(key: string): ExpressionMediaPayload | null {
	const hit = responseCache.get(key);
	if (!hit) {
		return null;
	}
	if (Date.now() > hit.expiresAt) {
		responseCache.delete(key);
		return null;
	}
	return hit.payload;
}

function setCachedPayload(key: string, payload: ExpressionMediaPayload, ttlMs: number) {
	if (responseCache.size >= MAX_CACHE_ENTRIES) {
		const oldestKey = responseCache.keys().next().value;
		if (oldestKey) {
			responseCache.delete(oldestKey);
		}
	}
	responseCache.set(key, { expiresAt: Date.now() + ttlMs, payload });
}

type KlipyMediaVariant = {
	url?: string;
	width?: number;
	height?: number;
};

type KlipyItem = {
	id?: string | number;
	slug?: string;
	title?: string;
	file?: {
		hd?: { gif?: KlipyMediaVariant; webp?: KlipyMediaVariant };
		md?: { gif?: KlipyMediaVariant; webp?: KlipyMediaVariant };
		sm?: { gif?: KlipyMediaVariant; webp?: KlipyMediaVariant };
	};
};

type KlipyResponse = {
	data?: {
		data?: KlipyItem[];
	};
};

type GiphyItem = {
	id?: string;
	title?: string;
	images?: {
		fixed_width?: { url?: string; width?: string; height?: string };
		downsized?: { url?: string; width?: string; height?: string };
		original?: { url?: string; width?: string; height?: string };
	};
};

type GiphyResponse = {
	data?: GiphyItem[];
};

function mapKlipyItems(raw: KlipyItem[], fallbackTitle: string): MediaExpressionEntry[] {
	const mapped: MediaExpressionEntry[] = [];
	for (const item of raw) {
		const full =
			item.file?.md?.webp ??
			item.file?.md?.gif ??
			item.file?.hd?.webp ??
			item.file?.hd?.gif ??
			item.file?.sm?.webp ??
			item.file?.sm?.gif;
		const preview =
			item.file?.sm?.webp ??
			item.file?.sm?.gif ??
			item.file?.md?.webp ??
			full;
		if (!full?.url || !preview?.url) {
			continue;
		}
		mapped.push({
			id: String(item.id ?? item.slug ?? full.url),
			title: item.title || fallbackTitle,
			url: full.url,
			previewUrl: preview.url,
			width: Number(full.width ?? 240),
			height: Number(full.height ?? 180),
			provider: "klipy",
		});
	}
	return mapped;
}

function mapGiphyItems(raw: GiphyItem[], fallbackTitle: string): MediaExpressionEntry[] {
	const mapped: MediaExpressionEntry[] = [];
	for (const item of raw) {
		const full = item.images?.downsized ?? item.images?.original;
		const preview = item.images?.fixed_width ?? full;
		if (!full?.url || !preview?.url) {
			continue;
		}
		mapped.push({
			id: String(item.id ?? full.url),
			title: item.title || fallbackTitle,
			url: full.url,
			previewUrl: preview.url,
			width: Number(full.width ?? 240),
			height: Number(full.height ?? 180),
			provider: "giphy",
		});
	}
	return mapped;
}

async function fetchKlipyMedia(
	appKey: string,
	kind: MediaExpressionKind,
	query: string,
	limit: number,
): Promise<{ ok: boolean; items: MediaExpressionEntry[] }> {
	const base = `https://api.klipy.com/api/v1/${encodeURIComponent(appKey)}/${kind}`;
	const url = query
		? `${base}/search?q=${encodeURIComponent(query)}&per_page=${limit}`
		: `${base}/trending?per_page=${limit}`;
	try {
		const res = await fetch(url, { next: { revalidate: 180 } });
		if (!res.ok) {
			return { ok: false, items: [] };
		}
		const body = (await res.json()) as KlipyResponse;
		return {
			ok: true,
			items: mapKlipyItems(body.data?.data ?? [], kind === "stickers" ? "Sticker" : "GIF"),
		};
	} catch {
		return { ok: false, items: [] };
	}
}

async function fetchGiphyMedia(
	apiKey: string,
	kind: MediaExpressionKind,
	query: string,
	limit: number,
): Promise<{ ok: boolean; items: MediaExpressionEntry[] }> {
	const base = `https://api.giphy.com/v1/${kind}`;
	const url = query
		? `${base}/search?api_key=${encodeURIComponent(apiKey)}&q=${encodeURIComponent(query)}&limit=${limit}&rating=pg-13`
		: `${base}/trending?api_key=${encodeURIComponent(apiKey)}&limit=${limit}&rating=pg-13`;
	try {
		const res = await fetch(url, { next: { revalidate: 180 } });
		if (!res.ok) {
			return { ok: false, items: [] };
		}
		const body = (await res.json()) as GiphyResponse;
		return {
			ok: true,
			items: mapGiphyItems(body.data ?? [], kind === "stickers" ? "Sticker" : "GIF"),
		};
	} catch {
		return { ok: false, items: [] };
	}
}

export async function resolveExpressionMediaRequest(params: {
	kind: MediaExpressionKind;
	mode: string;
	query: string;
	limit: number;
}): Promise<ExpressionMediaPayload> {
	const klipyKey = process.env.KLIPY_APP_KEY?.trim();
	const giphyKey = process.env.GIPHY_API_KEY?.trim();

	if (!klipyKey && !giphyKey) {
		return {
			configured: false,
			ok: false,
			provider: "none",
			items: [],
			categories: [],
		};
	}

	const cacheKey = `${params.kind}:${params.mode}:${params.query.toLowerCase()}:${params.limit}`;
	const cached = getCachedPayload(cacheKey);
	if (cached) {
		return cached;
	}

	const inflight = inflightRequests.get(cacheKey);
	if (inflight) {
		return inflight;
	}

	const provider = klipyKey ? ("klipy" as const) : ("giphy" as const);
	const fetchBatch = (q: string, n: number) =>
		klipyKey
			? fetchKlipyMedia(klipyKey, params.kind, q, n)
			: fetchGiphyMedia(giphyKey!, params.kind, q, n);

	const job = (async (): Promise<ExpressionMediaPayload> => {
		if (params.mode === "categories") {
			const defs = DEFAULT_MEDIA_CATEGORIES[params.kind];
			const results = await Promise.all(
				defs.map(async (cat) => {
					const res = await fetchBatch(cat.query, 1);
					return {
						ok: res.ok,
						category: {
							...cat,
							previewUrl: res.items[0]?.previewUrl ?? "",
						},
					};
				}),
			);
			const anyOk = results.some((r) => r.ok && Boolean(r.category.previewUrl));
			const payload: ExpressionMediaPayload = {
				configured: true,
				ok: anyOk,
				provider,
				items: [],
				categories: anyOk ? results.map((r) => r.category) : [],
			};
			if (anyOk) {
				setCachedPayload(cacheKey, payload, CATEGORIES_CACHE_TTL_MS);
			}
			return payload;
		}

		const result = await fetchBatch(params.query, params.limit);
		const payload: ExpressionMediaPayload = {
			configured: true,
			ok: result.ok,
			provider,
			items: result.items,
			categories: [],
		};
		if (result.ok) {
			setCachedPayload(cacheKey, payload, LIST_CACHE_TTL_MS);
		}
		return payload;
	})();

	inflightRequests.set(cacheKey, job);
	try {
		return await job;
	} finally {
		inflightRequests.delete(cacheKey);
	}
}
