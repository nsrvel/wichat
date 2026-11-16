"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useSyncExternalStore } from "react";
import { WorkspaceLoader } from "@/components/layout/workspace-loader";
import {
	getGatewayAttempt,
	getGatewayStatus,
	onGatewayRecovered,
	subscribeGatewayStatus,
} from "@/integrations/realtime/gateway-connection";

const CONNECTION_PROBLEMS_ATTEMPTS = 2;
const getServerStatus = () => "connected" as const;
const getServerAttempt = () => 0;

export function GatewayConnectionWatcher() {
	const queryClient = useQueryClient();
	const status = useSyncExternalStore(subscribeGatewayStatus, getGatewayStatus, getServerStatus);
	const attempt = useSyncExternalStore(subscribeGatewayStatus, getGatewayAttempt, getServerAttempt);

	useEffect(() => {
		return onGatewayRecovered(() => {
			void queryClient.invalidateQueries();
		});
	}, [queryClient]);

	if (status === "reconnecting" && attempt >= CONNECTION_PROBLEMS_ATTEMPTS) {
		return (
			<div className="fixed inset-0 z-[90] flex flex-col bg-background">
				<WorkspaceLoader showLabel labelKey="shell.connection.problems" />
			</div>
		);
	}

	return null;
}
