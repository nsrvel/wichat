import { env } from "@/lib/env";

function objectPublicBase(): string {
	return env.NEXT_PUBLIC_OBJECT_PUBLIC_BASE_URL?.replace(/\/$/, "") ?? "";
}

/** Dev: load MinIO objects through Next rewrite (same origin as web). */
function devSameOriginObjectPath(href: string): string | null {
	if (process.env.NODE_ENV !== "development") {
		return null;
	}
	if (href.startsWith("/wichat-public/")) {
		return href;
	}
	try {
		const parsed = new URL(href);
		if (
			(parsed.hostname === "localhost" || parsed.hostname === "127.0.0.1") &&
			parsed.pathname.startsWith("/wichat-public/")
		) {
			return parsed.pathname;
		}
	} catch {
		// not absolute
	}
	return null;
}

/** Resolves group logo URLs from the API (absolute or legacy path-only). */
export function resolveGroupLogoUrl(logoUrl: string): string {
	let trimmed = logoUrl.trim();
	if (!trimmed) {
		return "";
	}

	if (!trimmed.startsWith("/") && !/^https?:\/\//i.test(trimmed) && !trimmed.startsWith("//")) {
		trimmed = `/${trimmed}`;
	}

	if (/^https?:\/\//i.test(trimmed)) {
		const devPath = devSameOriginObjectPath(trimmed);
		return devPath ?? trimmed;
	}
	if (trimmed.startsWith("//")) {
		const absolute = `https:${trimmed}`;
		const devPath = devSameOriginObjectPath(absolute);
		return devPath ?? absolute;
	}

	const base = objectPublicBase();
	if (!trimmed.startsWith("/") || !base) {
		const devPath = devSameOriginObjectPath(trimmed);
		return devPath ?? trimmed;
	}

	const normalizedBase = base.endsWith("/") ? base : `${base}/`;

	// API / storage key path: /groups/{id}/icon.jpg → {base}/groups/...
	if (trimmed.startsWith("/groups/")) {
		const joined = new URL(trimmed.slice(1), normalizedBase).href;
		const devPath = devSameOriginObjectPath(joined);
		return devPath ?? joined;
	}

	// Already includes bucket path (e.g. /wichat-public/groups/...)
	const basePath = new URL(normalizedBase).pathname.replace(/\/$/, "");
	if (basePath && trimmed.startsWith(`${basePath}/`)) {
		const joined =
			process.env.NODE_ENV === "development"
				? trimmed
				: new URL(`${trimmed.slice(basePath.length + 1)}`, normalizedBase).href;
		const devPath = devSameOriginObjectPath(joined);
		return devPath ?? joined;
	}

	const joined = new URL(trimmed.slice(1), normalizedBase).href;
	const devPath = devSameOriginObjectPath(joined);
	return devPath ?? joined;
}
