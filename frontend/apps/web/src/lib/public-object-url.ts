import { env } from "@/lib/env";

function objectPublicBase(): string {
	return env.NEXT_PUBLIC_OBJECT_PUBLIC_BASE_URL?.replace(/\/$/, "") ?? "";
}

function isLoopbackHost(hostname: string): boolean {
	return hostname === "localhost" || hostname === "127.0.0.1" || hostname === "[::1]";
}

/**
 * Extracts a same-origin `/wichat-public/...` path when `url` points to a local
 * MinIO/object-storage URL (e.g. `http://localhost:9000/wichat-public/...`).
 * Returns `null` when the URL is already relative or points to a non-local host.
 */
export function toSameOriginObjectPath(url: string): string | null {
	if (!/^https?:\/\//i.test(url)) {
		return null;
	}
	try {
		const parsed = new URL(url);
		if (!isLoopbackHost(parsed.hostname)) {
			return null;
		}
		if (parsed.pathname.startsWith("/wichat-public/")) {
			return `${parsed.pathname}${parsed.search}`;
		}
		const base = objectPublicBase();
		if (base) {
			const basePath = new URL(base).pathname.replace(/\/$/, "");
			if (basePath && parsed.pathname.startsWith(`${basePath}/`)) {
				return `${parsed.pathname}${parsed.search}`;
			}
		}
	} catch {
		return null;
	}
	return null;
}

function shouldPreferSameOriginProxy(): boolean {
	if (typeof window === "undefined") {
		return false;
	}
	return window.location.protocol === "https:" || !isLoopbackHost(window.location.hostname);
}

/**
 * Resolves public object storage URLs (group logos, banners, user avatars) from the API.
 * On HTTPS or non-loopback origins (e.g. Safari on a tunneled domain), local MinIO URLs
 * (`http://localhost:9000/wichat-public/...`) are rewritten to same-origin `/wichat-public/...`
 * so browsers do not block them as Mixed Content or unreachable loopback.
 */
export function resolvePublicObjectUrl(rawUrl: string): string {
	let trimmed = rawUrl.trim();
	if (!trimmed) {
		return "";
	}

	if (!trimmed.startsWith("/") && !/^https?:\/\//i.test(trimmed) && !trimmed.startsWith("//")) {
		trimmed = `/${trimmed}`;
	}

	if (/^https?:\/\//i.test(trimmed)) {
		if (shouldPreferSameOriginProxy()) {
			return toSameOriginObjectPath(trimmed) ?? trimmed;
		}
		return trimmed;
	}
	if (trimmed.startsWith("//")) {
		return `https:${trimmed}`;
	}

	const base = objectPublicBase();
	if (!trimmed.startsWith("/") || !base) {
		return trimmed;
	}

	const normalizedBase = base.endsWith("/") ? base : `${base}/`;

	// API / storage key path: /groups/{id}/icon.jpg or /users/{id}/avatar.jpg → {base}/...
	let resolved: string;
	if (trimmed.startsWith("/groups/") || trimmed.startsWith("/users/")) {
		resolved = new URL(trimmed.slice(1), normalizedBase).href;
	} else {
		// Already includes bucket path (e.g. /wichat-public/groups/...)
		const basePath = new URL(normalizedBase).pathname.replace(/\/$/, "");
		if (basePath && trimmed.startsWith(`${basePath}/`)) {
			resolved = new URL(trimmed.slice(basePath.length + 1), normalizedBase).href;
		} else {
			resolved = new URL(trimmed.slice(1), normalizedBase).href;
		}
	}

	if (shouldPreferSameOriginProxy()) {
		return toSameOriginObjectPath(resolved) ?? resolved;
	}
	return resolved;
}
