/** Decode JWT `exp` (seconds) for client-side refresh hints — no signature verify. */
export function accessTokenExpiresAtMs(token: string): number | null {
	const parts = token.split(".");
	if (parts.length < 2) {
		return null;
	}
	const segment = parts[1];
	if (!segment) {
		return null;
	}
	try {
		const base64 = segment.replace(/-/g, "+").replace(/_/g, "/");
		const padded = base64 + "=".repeat((4 - (base64.length % 4)) % 4);
		const json = JSON.parse(atob(padded)) as { exp?: unknown };
		return typeof json.exp === "number" ? json.exp * 1000 : null;
	} catch {
		return null;
	}
}

/** True when access JWT is missing, unreadable, or inside refresh skew of expiry. */
export function isAccessTokenStale(
	token: string | null | undefined,
	skewMs = 30_000,
): boolean {
	if (!token?.trim()) {
		return true;
	}
	const expMs = accessTokenExpiresAtMs(token);
	if (expMs === null) {
		return false;
	}
	return Date.now() + skewMs >= expMs;
}
