"use client";

import type { ReactNode } from "react";
import { usePersistedShellRailOpen } from "@/hooks/use-persisted-shell-rail-open";
import { ShellMainPane } from "@/components/layout/shell-main-pane";
import { shellChromeBodyClass } from "@/components/layout/shell-rail/shell-rail-aside";
import { ShellRailTrack } from "@/components/layout/shell-rail/shell-rail-track";
import { ShellChromeProvider } from "@/components/layout/shell-chrome-context";
import { ShellTopBar } from "@/components/layout/shell-top-bar";
import { WorkspacePanel } from "@/components/layout/workspace-panel";
import { ProductionChannelSidebar } from "@/features/channel/components/production-channel-sidebar";
import type { QuickViewId } from "@/lib/app-routes";
import { GroupOnboardingProvider } from "@/features/group/group-onboarding-provider";
import { CreateCategoryDialogProvider } from "@/features/channel/create-category-dialog-provider";
import { CreateChannelDialogProvider } from "@/features/channel/create-channel-dialog-provider";
import { GroupSettingsProvider } from "@/features/group-settings/group-settings-provider";
import { InviteDialogProvider } from "@/features/invite/invite-dialog-provider";
import { useRedeemInviteFromQuery } from "@/features/invite/hooks/use-redeem-invite";
import { ShellGroupRail } from "@/features/shell/shell-group-rail";
import { ShellMembershipGate } from "@/features/shell/shell-membership-gate";

export type ShellLayoutProps = {
	children?: ReactNode;
	groupId?: string;
	channelId?: string;
	quickView?: QuickViewId;
};

function ShellInviteEffects() {
	useRedeemInviteFromQuery();
	return null;
}

export function ShellLayout({ children, groupId, channelId, quickView }: ShellLayoutProps) {
	const [railOpen, setRailOpen] = usePersistedShellRailOpen();

	const inWorkspace = Boolean(groupId && (channelId || quickView));

	const chrome = (
		<div className="flex h-full min-h-0 w-full flex-col overflow-hidden bg-background text-foreground">
			<ShellChromeProvider railOpen={railOpen}>
				<div className={shellChromeBodyClass(inWorkspace ? "relative" : undefined)}>
					<ShellTopBar railOpen={railOpen} onRailOpenChange={setRailOpen} />

					<ShellRailTrack
						open={railOpen}
						rail={
							<ShellGroupRail open={railOpen} activeGroupId={groupId} activeChannelId={channelId} />
						}
					>
						<ShellMainPane railOpen={railOpen}>
							{inWorkspace && groupId ? (
								<WorkspacePanel
									groupId={groupId}
									channelId={channelId}
									quickView={quickView}
									sidebar={
										<ProductionChannelSidebar
											groupId={groupId}
											activeChannelId={channelId ?? ""}
											activeQuickView={quickView}
											panelResizeHandle
										/>
									}
								>
									<ShellMembershipGate groupId={groupId}>{children}</ShellMembershipGate>
								</WorkspacePanel>
							) : (
								children
							)}
						</ShellMainPane>
					</ShellRailTrack>
				</div>
			</ShellChromeProvider>
		</div>
	);

	return (
		<GroupOnboardingProvider>
			<InviteDialogProvider>
				<GroupSettingsProvider>
					<CreateCategoryDialogProvider>
						<CreateChannelDialogProvider>
							<ShellInviteEffects />
							{chrome}
						</CreateChannelDialogProvider>
					</CreateCategoryDialogProvider>
				</GroupSettingsProvider>
			</InviteDialogProvider>
		</GroupOnboardingProvider>
	);
}
