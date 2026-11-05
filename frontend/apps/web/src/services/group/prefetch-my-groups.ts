import type { QueryClient } from "@tanstack/react-query";
import { myGroupsQueryOpts } from "@/services/group/group.query";

export function prefetchMyGroups(queryClient: QueryClient) {
	return queryClient.prefetchQuery(myGroupsQueryOpts());
}
