"use client";

import { getQueryClient } from "@/integrations/tanstack-query/get-query-client";
import { useMounted } from "@/hooks/use-mounted";
import { TooltipProvider } from "@/components/ui/tooltip";
import { GatewayConnectionWatcher } from "@/features/shell/gateway-connection-watcher";
import { VoiceSessionRoot } from "@/features/voice/voice-session-root";
import { suppressDefaultContextMenuUnlessAllowed } from "@/lib/context-menu-chrome";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";

export function AppProviders({ children }: { children: React.ReactNode }) {
	const queryClient = getQueryClient();
	const mounted = useMounted();

	return (
		<QueryClientProvider client={queryClient}>
			<GatewayConnectionWatcher />
			<TooltipProvider>
				<VoiceSessionRoot>
					<div
						className="flex min-h-0 min-w-0 flex-1 flex-col select-none"
						onContextMenuCapture={suppressDefaultContextMenuUnlessAllowed}
					>
						{children}
					</div>
				</VoiceSessionRoot>
			</TooltipProvider>
			{mounted ? <ReactQueryDevtools initialIsOpen={false} buttonPosition="bottom-right" /> : null}
		</QueryClientProvider>
	);
}
