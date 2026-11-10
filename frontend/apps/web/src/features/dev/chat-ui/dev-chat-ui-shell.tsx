"use client";

import { cloneElement, isValidElement, useState, type ReactElement, type ReactNode } from "react";
import { usePersistedShellRailOpen } from "@/hooks/use-persisted-shell-rail-open";
import { ChannelSidebar } from "@/components/layout/channel-sidebar";
import { IconRail } from "@/components/layout/shell-rail/icon-rail";
import { ShellMainPane } from "@/components/layout/shell-main-pane";
import { shellChromeBodyClass } from "@/components/layout/shell-rail/shell-rail-aside";
import { ShellRailTrack } from "@/components/layout/shell-rail/shell-rail-track";
import { WorkspacePanel } from "@/components/layout/workspace-panel";
import { ShellNavigationProvider } from "@/components/layout/shell-navigation";
import { ShellChromeProvider } from "@/components/layout/shell-chrome-context";
import { ShellTopBar } from "@/components/layout/shell-top-bar";
import type { ShellMainChildProps } from "@/components/layout/app-shell";

/** Full mock chat chrome — only for `/dev/chat-ui`. */
export function DevChatUiShell({ children }: { children: ReactNode }) {
	const [sidebarOpen, setSidebarOpen] = useState(false);
	const [railOpen, setRailOpen] = usePersistedShellRailOpen();
	const main = isValidElement(children)
		? cloneElement(children as ReactElement<ShellMainChildProps>, {
				onOpenSidebar: () => setSidebarOpen(true),
			})
		: children;

	return (
		<ShellNavigationProvider>
			<div className="flex h-full min-h-0 w-full flex-col overflow-hidden bg-background text-foreground">
				<ShellChromeProvider railOpen={railOpen}>
					<div className={shellChromeBodyClass("relative")}>
						<ShellTopBar railOpen={railOpen} onRailOpenChange={setRailOpen} />

						<ShellRailTrack open={railOpen} rail={<IconRail open={railOpen} />}>
							<ShellMainPane railOpen={railOpen}>
								<WorkspacePanel onOpenSidebar={() => setSidebarOpen(true)}>{main}</WorkspacePanel>
							</ShellMainPane>
						</ShellRailTrack>

						{sidebarOpen ? (
							<div className="fixed inset-0 z-40 lg:hidden">
								<button
									type="button"
									className="absolute inset-0 bg-black/30"
									aria-label="Close channels"
									onClick={() => setSidebarOpen(false)}
								/>
								<div className="absolute top-0 left-0 h-full shadow-2xl">
									<ChannelSidebar resizable={false} />
								</div>
							</div>
						) : null}
					</div>
				</ShellChromeProvider>
			</div>
		</ShellNavigationProvider>
	);
}
