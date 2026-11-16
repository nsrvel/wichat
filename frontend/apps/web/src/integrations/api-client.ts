import { createApiClient } from "@wichat/api-client";
import { createSessionTokenStorage } from "@wichat/api-client/token-storage";
import {
	notifyGatewayReachable,
	notifyGatewayUnreachable,
} from "@/integrations/realtime/gateway-connection";
import { env } from "@/lib/env";

const storage = createSessionTokenStorage();

export const { api, ensureAccessToken, getAccessToken, setAccessToken } = createApiClient({
	getBaseUrl: () => `${env.NEXT_PUBLIC_API_URL}/api/v1`,
	credentials: "include",
	storage,
	onGatewayUnreachable: notifyGatewayUnreachable,
	onGatewayReachable: notifyGatewayReachable,
});
