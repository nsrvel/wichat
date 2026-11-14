"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { WorkspaceLoader } from "@/components/layout/workspace-loader";
import { useMounted } from "@/hooks/use-mounted";
import { bootstrapApp, Q_APP_BOOTSTRAP } from "@/services/auth/bootstrap-app";

export function AuthGate({ children }: { children: React.ReactNode }) {
	const router = useRouter();
	const queryClient = useQueryClient();
	const mounted = useMounted();
	const bootstrap = useQuery({
		queryKey: [Q_APP_BOOTSTRAP],
		queryFn: () => bootstrapApp(queryClient),
		enabled: mounted,
		staleTime: 60_000,
		retry: false,
	});

	useEffect(() => {
		if (!mounted || bootstrap.isPending) {
			return;
		}
		if (!bootstrap.data) {
			router.replace("/sign-in");
		}
	}, [bootstrap.data, bootstrap.isPending, mounted, router]);

	if (!mounted || bootstrap.isPending) {
		return <WorkspaceLoader className="min-h-[50vh]" />;
	}

	if (!bootstrap.data) {
		return <WorkspaceLoader className="min-h-[50vh]" />;
	}

	return children;
}
