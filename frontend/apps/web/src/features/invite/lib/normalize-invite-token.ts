/** Extracts invite token from a raw code or pasted invite URL. */
export function normalizeInviteToken(raw: string): string {
	const trimmed = raw.trim();
	if (!trimmed) {
		return "";
	}

	const fromPath = extractInvitePathSegment(trimmed);
	if (fromPath) {
		return fromPath;
	}

	try {
		if (trimmed.includes("://")) {
			const url = new URL(trimmed);
			const fromQuery = url.searchParams.get("invite") ?? url.searchParams.get("token");
			if (fromQuery?.trim()) {
				return fromQuery.trim();
			}
			const fromUrlPath = extractInvitePathSegment(url.pathname);
			if (fromUrlPath) {
				return fromUrlPath;
			}
		}
		if (trimmed.startsWith("/")) {
			const url = new URL(trimmed, "https://wichat.local");
			const fromQuery = url.searchParams.get("invite") ?? url.searchParams.get("token");
			if (fromQuery?.trim()) {
				return fromQuery.trim();
			}
		}
	} catch {
		// fall through
	}

	return trimmed;
}

function extractInvitePathSegment(pathOrUrl: string): string {
	const marker = "/invite/";
	const idx = pathOrUrl.indexOf(marker);
	if (idx < 0) {
		return "";
	}
	const rest = pathOrUrl.slice(idx + marker.length);
	const seg = rest.split(/[?#]/)[0]?.trim();
	return seg ? decodeURIComponent(seg) : "";
}
