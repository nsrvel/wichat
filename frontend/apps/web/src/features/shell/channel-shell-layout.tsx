"use client";

import type { ReactNode } from "react";
import { ShellNavigationProvider } from "@/components/layout/shell-navigation";
import {
	ChannelNavigationProvider,
	useChannelNavigation,
} from "@/features/channel/channel-navigation-provider";
import { GroupRouteGuard } from "@/features/channel/group-route-guard";
import { ShellLayout } from "@/features/shell/shell-layout";

type ChannelShellLayoutProps = {
	children: ReactNode;
};

function ChannelShellLayoutInner({ children }: ChannelShellLayoutProps) {
	const { groupId, channelId, quickView } = useChannelNavigation();

	const shell = (
		<ShellLayout groupId={groupId} channelId={channelId} quickView={quickView}>
			{children}
		</ShellLayout>
	);

	if (!groupId) {
		return shell;
	}

	return (
		<GroupRouteGuard groupId={groupId}>
			<ShellNavigationProvider>{shell}</ShellNavigationProvider>
		</GroupRouteGuard>
	);
}

/** One shell for all `/channel/*` routes — rail does not remount when switching groups. */
export function ChannelShellLayout({ children }: ChannelShellLayoutProps) {
	return (
		<ChannelNavigationProvider>
			<ChannelShellLayoutInner>{children}</ChannelShellLayoutInner>
		</ChannelNavigationProvider>
	);
}
