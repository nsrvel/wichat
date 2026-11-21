import type { QueryClient } from "@tanstack/react-query";
import { Q_MY_GROUPS } from "@/services/group/group.key";

/** After a real outage: refresh rail group list only — members stay on HTTP seed + WS presence patches. */
export function invalidateAfterGatewayRecover(queryClient: QueryClient): void {
	void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
}
