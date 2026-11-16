"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useRef, useSyncExternalStore } from "react";
import { WorkspaceLoader } from "@/components/layout/workspace-loader";
import { WorkspaceReadyGate } from "@/features/shell/workspace-ready-gate";
import { useMounted } from "@/hooks/use-mounted";
import {
	getGatewayStatus,
	subscribeGatewayStatus,
} from "@/integrations/realtime/gateway-connection";
import { bootstrapApp, Q_APP_BOOTSTRAP } from "@/services/auth/bootstrap-app";

const getServerSnapshot = () => "connected" as const;

export function AuthGate({ children }: { children: React.ReactNode }) {
	const router = useRouter();
	const pathname = usePathname();
	const initialPathnameRef = useRef(pathname);
	const queryClient = useQueryClient();
	const mounted = useMounted();
	const gatewayStatus = useSyncExternalStore(
		subscribeGatewayStatus,
		getGatewayStatus,
		getServerSnapshot,
	);
	const bootstrap = useQuery({
		queryKey: [Q_APP_BOOTSTRAP],
		queryFn: () => bootstrapApp(queryClient, initialPathnameRef.current),
		enabled: mounted,
		staleTime: 60_000,
		retry: false,
	});

	useEffect(() => {
		if (!mounted || bootstrap.isPending || bootstrap.isError || gatewayStatus === "reconnecting") {
			return;
		}
		if (!bootstrap.data) {
			router.replace("/sign-in");
		}
	}, [bootstrap.data, bootstrap.isError, bootstrap.isPending, gatewayStatus, mounted, router]);

	if (!mounted || bootstrap.isPending || bootstrap.isError || gatewayStatus === "reconnecting") {
		if (bootstrap.data) {
			return <WorkspaceReadyGate>{children}</WorkspaceReadyGate>;
		}
		return <WorkspaceLoader className="min-h-[50vh]" />;
	}

	if (!bootstrap.data) {
		return <WorkspaceLoader className="min-h-[50vh]" />;
	}

	return <WorkspaceReadyGate>{children}</WorkspaceReadyGate>;
}
