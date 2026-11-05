export type TokenStorage = {
	get: () => string | null;
	set: (token: string | null) => void;
};

export function createMemoryTokenStorage(): TokenStorage {
	let accessToken: string | null = null;
	return {
		get: () => accessToken,
		set: (token) => {
			accessToken = token;
		},
	};
}

/** Survives full page reload in the same tab — avoids refresh-token rotation on every navigation. */
export function createSessionTokenStorage(
	storageKey = "wichat_access_token",
): TokenStorage {
	let cached: string | null | undefined;

	const readPersisted = (): string | null => {
		if (typeof sessionStorage === "undefined") {
			return null;
		}
		try {
			return sessionStorage.getItem(storageKey);
		} catch {
			return null;
		}
	};

	return {
		get: () => {
			if (cached !== undefined) {
				return cached;
			}
			cached = readPersisted();
			return cached;
		},
		set: (token) => {
			cached = token;
			if (typeof sessionStorage === "undefined") {
				return;
			}
			try {
				if (token) {
					sessionStorage.setItem(storageKey, token);
				} else {
					sessionStorage.removeItem(storageKey);
				}
			} catch {
				// private mode / quota — memory cache still works until reload
			}
		},
	};
}

let refreshInFlight: Promise<string | null> | null = null;

export function refreshAccessToken(
	refresh: () => Promise<string | null>,
): Promise<string | null> {
	if (!refreshInFlight) {
		refreshInFlight = refresh().finally(() => {
			refreshInFlight = null;
		});
	}

	return refreshInFlight;
}
