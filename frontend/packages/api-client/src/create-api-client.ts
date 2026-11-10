import { SessionSchema } from "@wichat/contracts";
import { ApiError } from "./api-error";
import { refreshAccessToken, type TokenStorage } from "./token-storage";

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
};

export function createApiClient({
	getBaseUrl,
	credentials = "include",
	storage,
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
		try {
			return await fetch(`${getBaseUrl()}${path}`, {
				...init,
				credentials,
			});
		} catch {
			throw new ApiError("error.network");
		}
	}

	async function refreshAccess(): Promise<string | null> {
		return refreshAccessToken(async () => {
			try {
				const res = await fetchApi("/auth/refresh", { method: "POST" });
				if (!res.ok) {
					setAccessToken(null);
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
				setAccessToken(null);
				return null;
			}
		});
	}

	async function ensureAccessToken(): Promise<string | null> {
		if (getAccessToken()) {
			return getAccessToken();
		}
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
		async get<T>(path: string, schema: Schema<T>): Promise<T> {
			return send(path, schema, { method: "GET" }, true);
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

		async delete<T>(path: string, schema: Schema<T>): Promise<T> {
			return send(path, schema, { method: "DELETE" }, true);
		},

		async postForm<T>(
			path: string,
			schema: Schema<T>,
			formData: FormData,
		): Promise<T> {
			return send(path, schema, { method: "POST", body: formData }, true);
		},
	};

	return { api, ensureAccessToken, getAccessToken, setAccessToken };
}
