"use client";

import type { ReactNode } from "react";
import { usePathname } from "next/navigation";
import { useMemo } from "react";
import { ShellNavigationProvider } from "@/components/layout/shell-navigation";
import { GroupRouteGuard } from "@/features/channel/group-route-guard";
import { parseChannelRoute } from "@/features/shell/lib/parse-channel-route";
import { ShellLayout } from "@/features/shell/shell-layout";

type ChannelShellLayoutProps = {
	children: ReactNode;
};

/** One shell for all `/channel/*` routes — rail does not remount when switching groups. */
export function ChannelShellLayout({ children }: ChannelShellLayoutProps) {
	const pathname = usePathname();
	const target = useMemo(() => parseChannelRoute(pathname), [pathname]);

	const shell = (
		<ShellLayout groupId={target.groupId} channelId={target.channelId} quickView={target.quickView}>
			{children}
		</ShellLayout>
	);

	if (!target.groupId) {
		return shell;
	}

	return (
		<GroupRouteGuard groupId={target.groupId}>
			<ShellNavigationProvider>{shell}</ShellNavigationProvider>
		</GroupRouteGuard>
	);
}
