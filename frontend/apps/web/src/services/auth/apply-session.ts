import type { Session } from "@wichat/contracts/auth";
import type { QueryClient } from "@tanstack/react-query";
import { Q_INSTANCE_POLICY, Q_SESSION } from "@/services/auth/auth.key";
import { bootstrapApp, Q_APP_BOOTSTRAP } from "@/services/auth/bootstrap-app";
import { prefetchMyGroups } from "@/services/group/prefetch-my-groups";

/** After OTP/password auth, sync session cache and app gate so sign-in → home does not loop. */
export function applySession(queryClient: QueryClient, session: Session): void {
	queryClient.setQueryData([Q_SESSION], session.user);
	queryClient.setQueryData([Q_APP_BOOTSTRAP], session.user);
	void queryClient.invalidateQueries({ queryKey: [Q_INSTANCE_POLICY] });
	void prefetchMyGroups(queryClient);
	void queryClient.fetchQuery({
		queryKey: [Q_APP_BOOTSTRAP],
		queryFn: () => bootstrapApp(queryClient),
	});
}
