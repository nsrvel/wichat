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
