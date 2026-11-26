import { SessionSchema } from "@wichat/contracts/auth";
import { ApiError } from "./api-error.ts";
import { isAccessTokenStale } from "./jwt-access.ts";
import { refreshAccessToken, type TokenStorage } from "./token-storage.ts";

type Schema<T> = { parse: (input: unknown) => T };

const noRefresh = new Set([
	"/auth/login",
	"/auth/login/verify",
	"/auth/register",
	"/auth/register/verify",
	"/auth/setup",
	"/auth/refresh",
	"/auth/logout",
]);

export type CreateApiClientOptions = {
	getBaseUrl: () => string;
	credentials?: RequestCredentials;
	storage: TokenStorage;
	onGatewayUnreachable?: () => void;
	onGatewayReachable?: () => void;
};

export function createApiClient({
	getBaseUrl,
	credentials = "include",
	storage,
	onGatewayUnreachable,
	onGatewayReachable,
}: CreateApiClientOptions) {
	const getAccessToken = () => storage.get();
	const setAccessToken = (token: string | null) => storage.set(token);

	async function parseResponse<T>(
		res: Response,
		schema: Schema<T>,
	): Promise<T> {
		const json: unknown = await res.json().catch(() => null);
		if (!res.ok) {
			throw ApiError.fromEnvelope(json);
		}
		if (!json || typeof json !== "object" || !("data" in json)) {
			throw new ApiError("error.invalid_response");
		}
		try {
			return schema.parse((json as { data: unknown }).data);
		} catch {
			throw new ApiError("error.invalid_response");
		}
	}

	async function fetchApi(path: string, init: RequestInit): Promise<Response> {
		let res: Response;
		try {
			res = await fetch(`${getBaseUrl()}${path}`, {
				...init,
				credentials,
			});
		} catch (err) {
			if (err instanceof DOMException && err.name === "AbortError") {
				throw err;
			}
			onGatewayUnreachable?.();
			throw new ApiError("error.network");
		}
		if (res.status === 502 || res.status === 503 || res.status === 504) {
			onGatewayUnreachable?.();
		} else {
			onGatewayReachable?.();
		}
		return res;
	}

	async function refreshAccess(): Promise<string | null> {
		return refreshAccessToken(async () => {
			try {
				const res = await fetchApi("/auth/refresh", { method: "POST" });
				if (!res.ok) {
					if (res.status === 401 || res.status === 403) {
						setAccessToken(null);
					}
					return null;
				}
				const json: unknown = await res.json().catch(() => null);
				const data =
					json && typeof json === "object" && "data" in json
						? (json as { data: unknown }).data
						: null;
				const parsed = SessionSchema.safeParse(data);
				if (!parsed.success) {
					setAccessToken(null);
					return null;
				}
				setAccessToken(parsed.data.access_token);
				return parsed.data.access_token;
			} catch {
				// Preserve existing access token on transient network/gateway errors
				return null;
			}
		});
	}

	async function ensureAccessToken(): Promise<string | null> {
		const current = getAccessToken();
		if (current && !isAccessTokenStale(current)) {
			return current;
		}
		try {
			const refreshed = await refreshAccess();
			if (refreshed) {
				return refreshed;
			}
			return current && !isAccessTokenStale(current) ? current : null;
		} catch {
			return current && !isAccessTokenStale(current) ? current : null;
		}
	}

	async function forceRefreshAccessToken(): Promise<string | null> {
		try {
			return await refreshAccess();
		} catch {
			return null;
		}
	}

	async function send<T>(
		path: string,
		schema: Schema<T>,
		init: RequestInit,
		allowRefresh: boolean,
	): Promise<T> {
		const headers = new Headers(init.headers);
		const token = getAccessToken();
		if (token) {
			headers.set("Authorization", `Bearer ${token}`);
		}

		const res = await fetchApi(path, { ...init, headers });

		if (res.status === 401 && allowRefresh && !noRefresh.has(path)) {
			const next = await refreshAccess();
			if (next) {
				return send(path, schema, init, false);
			}
		}

		return parseResponse(res, schema);
	}

	const api = {
		async get<T>(
			path: string,
			schema: Schema<T>,
			init?: RequestInit,
		): Promise<T> {
			return send(path, schema, { method: "GET", ...init }, true);
		},

		async post<T>(path: string, schema: Schema<T>, body: unknown): Promise<T> {
			return send(
				path,
				schema,
				{
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(body),
				},
				true,
			);
		},

		async put<T>(path: string, schema: Schema<T>, body: unknown): Promise<T> {
			return send(
				path,
				schema,
				{
					method: "PUT",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(body),
				},
				true,
			);
		},

		async patch<T>(path: string, schema: Schema<T>, body: unknown): Promise<T> {
			return send(
				path,
				schema,
				{
					method: "PATCH",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(body),
				},
				true,
			);
		},

		async delete<T>(
			path: string,
			schema: Schema<T>,
			body?: unknown,
		): Promise<T> {
			return send(
				path,
				schema,
				body === undefined
					? { method: "DELETE" }
					: {
							method: "DELETE",
							headers: { "Content-Type": "application/json" },
							body: JSON.stringify(body),
						},
				true,
			);
		},

		async postForm<T>(
			path: string,
			schema: Schema<T>,
			formData: FormData,
		): Promise<T> {
			return send(path, schema, { method: "POST", body: formData }, true);
		},
	};

	return {
		api,
		ensureAccessToken,
		forceRefreshAccessToken,
		getAccessToken,
		setAccessToken,
	};
}
