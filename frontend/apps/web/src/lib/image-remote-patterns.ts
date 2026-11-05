import type { RemotePattern } from "next/dist/shared/lib/image-config";

/** Builds Next.js image remotePatterns from public env URLs (object storage, etc.). */
export function remotePatternsFromPublicUrls(urls: (string | undefined)[]): RemotePattern[] {
	const patterns: RemotePattern[] = [];
	const seen = new Set<string>();

	for (const raw of urls) {
		const trimmed = raw?.trim();
		if (!trimmed) {
			continue;
		}
		try {
			const parsed = new URL(trimmed);
			const protocol = parsed.protocol.replace(":", "") as "http" | "https";
			if (protocol !== "http" && protocol !== "https") {
				continue;
			}
			const port = parsed.port;
			const pathnameBase = parsed.pathname.replace(/\/$/, "") || "";
			const pathname = `${pathnameBase}/**`;
			const key = `${protocol}://${parsed.hostname}:${port}${pathname}`;
			if (seen.has(key)) {
				continue;
			}
			seen.add(key);
			patterns.push({
				protocol,
				hostname: parsed.hostname,
				port,
				pathname,
			});
		} catch {
			// ignore invalid env values
		}
	}

	return patterns;
}
