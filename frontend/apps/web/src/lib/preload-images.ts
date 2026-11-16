const DEFAULT_PRELOAD_TIMEOUT_MS = 1500;
const DEFAULT_MAX_PRELOAD_IMAGES = 32;
const MAX_RETAINED_IMAGES = 64;

/** Keep decoded HTMLImageElement instances alive in memory so browser does not GC bitmaps before mount. */
const retainedImages = new Map<string, HTMLImageElement>();

function retainImage(src: string, img: HTMLImageElement) {
	if (retainedImages.has(src)) {
		retainedImages.delete(src);
	} else if (retainedImages.size >= MAX_RETAINED_IMAGES) {
		const oldestKey = retainedImages.keys().next().value;
		if (oldestKey !== undefined) {
			retainedImages.delete(oldestKey);
		}
	}
	retainedImages.set(src, img);
}

function preloadSingleImage(src: string): Promise<void> {
	const existing = retainedImages.get(src);
	if (existing?.complete && existing.naturalWidth > 0) {
		return Promise.resolve();
	}

	return new Promise((resolve) => {
		const img = new Image();
		img.decoding = "sync";

		const finish = () => {
			img.onload = null;
			img.onerror = null;
			resolve();
		};

		img.onload = () => {
			retainImage(src, img);
			if (typeof img.decode === "function") {
				void img
					.decode()
					.catch(() => undefined)
					.finally(finish);
				return;
			}
			finish();
		};
		img.onerror = finish;
		img.src = src;

		if (img.complete && img.naturalWidth > 0) {
			retainImage(src, img);
			finish();
		}
	});
}

/**
 * Best-effort parallel image preloader + decoder with deduplication, retained bitmap cache,
 * max item cap, and a hard timeout so slow or broken image URLs never block WorkspaceLoader.
 */
export async function preloadImages(
	urls: Iterable<string | null | undefined>,
	options?: {
		timeoutMs?: number;
		maxCount?: number;
	},
): Promise<void> {
	if (typeof window === "undefined" || typeof Image === "undefined") {
		return;
	}

	const maxCount = options?.maxCount ?? DEFAULT_MAX_PRELOAD_IMAGES;
	const timeoutMs = options?.timeoutMs ?? DEFAULT_PRELOAD_TIMEOUT_MS;

	const uniqueUrls: string[] = [];
	const seen = new Set<string>();
	for (const raw of urls) {
		const trimmed = raw?.trim();
		if (!trimmed || seen.has(trimmed)) {
			continue;
		}
		seen.add(trimmed);
		uniqueUrls.push(trimmed);
		if (uniqueUrls.length >= maxCount) {
			break;
		}
	}

	if (uniqueUrls.length === 0) {
		return;
	}

	let timer: ReturnType<typeof setTimeout> | undefined;
	const allLoaded = Promise.allSettled(uniqueUrls.map(preloadSingleImage));
	const timeout = new Promise<void>((resolve) => {
		timer = setTimeout(resolve, timeoutMs);
	});

	await Promise.race([allLoaded, timeout]);
	if (timer !== undefined) {
		clearTimeout(timer);
	}
}
