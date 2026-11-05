"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useMounted } from "@/hooks/use-mounted";
import * as authQuery from "@/services/auth/auth.query";
import { ensureAccessToken } from "@/integrations/api-client";
import { prefetchMyGroups } from "@/services/group/prefetch-my-groups";

export function AuthGate({ children }: { children: React.ReactNode }) {
	const router = useRouter();
	const queryClient = useQueryClient();
	const mounted = useMounted();
	const session = useQuery({ ...authQuery.sessionQueryOpts(), enabled: mounted });

	useEffect(() => {
		if (!mounted) {
			return;
		}
		void ensureAccessToken().then((token) => {
			if (token) {
				void prefetchMyGroups(queryClient);
			}
		});
	}, [mounted, queryClient]);

	useEffect(() => {
		if (!mounted || session.isPending) {
			return;
		}
		if (!session.data) {
			router.replace("/sign-in");
			return;
		}
		void prefetchMyGroups(queryClient);
	}, [mounted, queryClient, router, session.data, session.isPending]);

	if (!mounted || session.isPending || !session.data) {
		return <div className="flex-1" suppressHydrationWarning />;
	}

	return children;
}
