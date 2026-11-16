"use client";

import { useQueryClient } from "@tanstack/react-query";
import { Loader2, WifiOff } from "lucide-react";
import { useEffect, useSyncExternalStore } from "react";
import { Button } from "@/components/ui/button";
import {
	getGatewayStatus,
	onGatewayRecovered,
	probeGatewayNow,
	subscribeGatewayStatus,
} from "@/integrations/realtime/gateway-connection";

const getServerSnapshot = () => "connected" as const;

export function GatewayConnectionWatcher() {
	const queryClient = useQueryClient();
	const status = useSyncExternalStore(subscribeGatewayStatus, getGatewayStatus, getServerSnapshot);

	useEffect(() => {
		return onGatewayRecovered(() => {
			void queryClient.invalidateQueries();
		});
	}, [queryClient]);

	if (status !== "reconnecting") {
		return null;
	}

	return (
		<div
			role="status"
			aria-live="polite"
			className="fixed top-3 left-1/2 z-50 flex -translate-x-1/2 items-center gap-2.5 rounded-full border border-amber-500/30 bg-background/95 px-3.5 py-1.5 text-xs font-medium text-foreground shadow-lg backdrop-blur"
		>
			<WifiOff className="h-3.5 w-3.5 text-amber-500" />
			<Loader2 className="h-3.5 w-3.5 animate-spin text-muted-foreground" />
			<span>Connection lost. Reconnecting to server...</span>
			<Button
				type="button"
				variant="ghost"
				size="sm"
				className="h-6 px-2 text-xs"
				onClick={() => void probeGatewayNow()}
			>
				Retry
			</Button>
		</div>
	);
}
