"use client";

import type { ReactNode } from "react";
import { usePersistedShellRailOpen } from "@/hooks/use-persisted-shell-rail-open";
import { ShellMainPane } from "@/components/layout/shell-main-pane";
import { shellChromeBodyClass } from "@/components/layout/shell-rail/shell-rail-aside";
import { ShellRailTrack } from "@/components/layout/shell-rail/shell-rail-track";
import { ShellChromeProvider } from "@/components/layout/shell-chrome-context";
import { ShellUserDock } from "@/components/layout/shell-user-dock";
import { ShellTopBar } from "@/components/layout/shell-top-bar";
import { WorkspacePanel } from "@/components/layout/workspace-panel";
import { ProductionChannelSidebar } from "@/features/channel/components/production-channel-sidebar";
import { GroupOnboardingProvider } from "@/features/group/group-onboarding-provider";
import { ShellGroupRail } from "@/features/shell/shell-group-rail";
import { ShellMembershipGate } from "@/features/shell/shell-membership-gate";

export type ShellLayoutProps = {
	children?: ReactNode;
	groupId?: string;
	channelId?: string;
};

export function ShellLayout({ children, groupId, channelId }: ShellLayoutProps) {
	const [railOpen, setRailOpen] = usePersistedShellRailOpen();

	const inChannel = Boolean(groupId && channelId);

	const chrome = (
		<div className="flex h-full min-h-0 w-full flex-col overflow-hidden bg-background text-foreground">
			<ShellChromeProvider railOpen={railOpen}>
				<div className={shellChromeBodyClass(inChannel ? "relative" : undefined)}>
					<ShellTopBar railOpen={railOpen} onRailOpenChange={setRailOpen} />

					<ShellRailTrack
						open={railOpen}
						rail={
							<ShellGroupRail open={railOpen} activeGroupId={groupId} activeChannelId={channelId} />
						}
					>
						<ShellMainPane railOpen={railOpen}>
							{inChannel && groupId && channelId ? (
								<WorkspacePanel
									sidebar={
										<ProductionChannelSidebar
											groupId={groupId}
											activeChannelId={channelId}
											panelResizeHandle
										/>
									}
								>
									{children}
								</WorkspacePanel>
							) : (
								children
							)}
						</ShellMainPane>
					</ShellRailTrack>

					{inChannel ? <ShellUserDock /> : null}
				</div>
			</ShellChromeProvider>
		</div>
	);

	return (
		<GroupOnboardingProvider>
			{groupId ? <ShellMembershipGate groupId={groupId}>{chrome}</ShellMembershipGate> : chrome}
		</GroupOnboardingProvider>
	);
}
