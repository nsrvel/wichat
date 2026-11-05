"use client";

import { getQueryClient } from "@/integrations/tanstack-query/get-query-client";
import { useMounted } from "@/hooks/use-mounted";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";

export function AppProviders({ children }: { children: React.ReactNode }) {
	const queryClient = getQueryClient();
	const mounted = useMounted();

	return (
		<QueryClientProvider client={queryClient}>
			<TooltipProvider>{children}</TooltipProvider>
			{mounted ? <ReactQueryDevtools initialIsOpen={false} buttonPosition="bottom-right" /> : null}
		</QueryClientProvider>
	);
}
