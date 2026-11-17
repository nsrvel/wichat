import { resolvePublicObjectUrl } from "@/lib/public-object-url";

/** Resolves group logo/banner URLs from the API (absolute or legacy path-only). */
export function resolveGroupLogoUrl(logoUrl: string): string {
	return resolvePublicObjectUrl(logoUrl);
}
