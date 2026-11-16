import { env } from "@/lib/env";

function objectPublicBase(): string {
	return env.NEXT_PUBLIC_OBJECT_PUBLIC_BASE_URL?.replace(/\/$/, "") ?? "";
}

/** Resolves group logo/banner URLs from the API (absolute or legacy path-only). */
export function resolveGroupLogoUrl(logoUrl: string): string {
	let trimmed = logoUrl.trim();
	if (!trimmed) {
		return "";
	}

	if (!trimmed.startsWith("/") && !/^https?:\/\//i.test(trimmed) && !trimmed.startsWith("//")) {
		trimmed = `/${trimmed}`;
	}

	if (/^https?:\/\//i.test(trimmed)) {
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

	// API / storage key path: /groups/{id}/icon.jpg → {base}/groups/...
	if (trimmed.startsWith("/groups/")) {
		return new URL(trimmed.slice(1), normalizedBase).href;
	}

	// Already includes bucket path (e.g. /wichat-public/groups/...)
	const basePath = new URL(normalizedBase).pathname.replace(/\/$/, "");
	if (basePath && trimmed.startsWith(`${basePath}/`)) {
		return new URL(`${trimmed.slice(basePath.length + 1)}`, normalizedBase).href;
	}

	return new URL(trimmed.slice(1), normalizedBase).href;
}
