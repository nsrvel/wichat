"use client";

import type { ReactNode } from "react";
import { ChannelRouteGuard } from "@/features/channel/channel-route-guard";
import { ShellNavigationProvider } from "@/components/layout/shell-navigation";
import { ShellLayout } from "@/features/shell/shell-layout";

type ChatShellLayoutProps = {
	groupId: string;
	channelId: string;
	children?: ReactNode;
};

/** Channel routes: production channel sidebar in shell; main pane empty until chat ships. */
export function ChatShellLayout({ groupId, channelId, children }: ChatShellLayoutProps) {
	return (
		<ChannelRouteGuard groupId={groupId} channelId={channelId}>
			<ShellNavigationProvider>
				<ShellLayout groupId={groupId} channelId={channelId}>
					{children}
				</ShellLayout>
			</ShellNavigationProvider>
		</ChannelRouteGuard>
	);
}
