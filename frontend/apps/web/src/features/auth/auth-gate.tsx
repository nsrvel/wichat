"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useMounted } from "@/hooks/use-mounted";
import * as authQuery from "@/services/auth/auth.query";

export function AuthGate({ children }: { children: React.ReactNode }) {
	const router = useRouter();
	const mounted = useMounted();
	const session = useQuery({ ...authQuery.sessionQueryOpts(), enabled: mounted });

	useEffect(() => {
		if (!mounted || session.isPending) {
			return;
		}
		if (!session.data) {
			router.replace("/sign-in");
		}
	}, [mounted, router, session.data, session.isPending]);

	if (!mounted || session.isPending || !session.data) {
		return <div className="flex-1" suppressHydrationWarning />;
	}

	return children;
}
