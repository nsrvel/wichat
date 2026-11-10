"use client";

import type { ReactNode } from "react";
import { usePathname } from "next/navigation";
import { useMemo } from "react";
import { ShellNavigationProvider } from "@/components/layout/shell-navigation";
import { GroupRouteGuard } from "@/features/channel/group-route-guard";
import { parseWorkspacePath } from "@/features/shell/lib/parse-workspace-path";
import { ShellLayout } from "@/features/shell/shell-layout";

type GroupWorkspaceShellLayoutProps = {
	groupId: string;
	children?: ReactNode;
};

/** One shell instance per group — avoids remounting the rail when switching channel vs quick views. */
export function GroupWorkspaceShellLayout({ groupId, children }: GroupWorkspaceShellLayoutProps) {
	const pathname = usePathname();
	const target = useMemo(() => parseWorkspacePath(pathname, groupId), [pathname, groupId]);

	return (
		<GroupRouteGuard groupId={groupId}>
			<ShellNavigationProvider>
				<ShellLayout groupId={groupId} channelId={target.channelId} quickView={target.quickView}>
					{children}
				</ShellLayout>
			</ShellNavigationProvider>
		</GroupRouteGuard>
	);
}
