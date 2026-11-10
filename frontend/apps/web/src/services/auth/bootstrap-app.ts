import type { User } from "@wichat/contracts";
import type { QueryClient } from "@tanstack/react-query";
import { ensureAccessToken } from "@/integrations/api-client";
import * as authQuery from "@/services/auth/auth.query";
import { myGroupsQueryOpts } from "@/services/group/group.query";

export const Q_APP_BOOTSTRAP = "app-bootstrap";

/** One refresh, then session + groups in parallel — avoids serial AuthGate stalls. */
export async function bootstrapApp(queryClient: QueryClient): Promise<User | null> {
	const token = await ensureAccessToken();
	if (!token) {
		return null;
	}

	const [user] = await Promise.all([
		queryClient.fetchQuery(authQuery.sessionQueryOpts()),
		queryClient.fetchQuery(myGroupsQueryOpts()),
	]);

	return user;
}
