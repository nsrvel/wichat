import { createApiClient, createMemoryTokenStorage } from "@wichat/api-client";
import { env } from "@/env";

const storage = createMemoryTokenStorage();

export const { api, ensureAccessToken, getAccessToken, setAccessToken } = createApiClient({
	getBaseUrl: () => `${env.NEXT_PUBLIC_API_URL}/api/v1`,
	credentials: "include",
	storage,
});
