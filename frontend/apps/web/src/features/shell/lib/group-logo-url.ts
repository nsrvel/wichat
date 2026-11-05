import { env } from "@/env";

/** Resolves group logo URLs from the API (absolute or legacy path-only). */
export function resolveGroupLogoUrl(logoUrl: string): string {
	const trimmed = logoUrl.trim();
	if (!trimmed) {
		return "";
	}
	if (/^https?:\/\//i.test(trimmed) || trimmed.startsWith("//")) {
		return trimmed;
	}
	if (trimmed.startsWith("/")) {
		const base = env.NEXT_PUBLIC_OBJECT_PUBLIC_BASE_URL?.replace(/\/$/, "");
		if (base) {
			return `${base}${trimmed}`;
		}
	}
	return trimmed;
}
