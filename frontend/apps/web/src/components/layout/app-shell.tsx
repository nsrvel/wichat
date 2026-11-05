"use client";

import type { ReactNode } from "react";

type AppShellProps = {
	children: ReactNode;
};

/** Production app wrapper — no chat chrome (mock shell lives under `/dev/chat-ui`). */
export function AppShell({ children }: AppShellProps) {
	return (
		<div className="flex h-full min-h-0 w-full flex-col overflow-hidden bg-background text-foreground">
			<div className="flex min-h-0 min-w-0 flex-1 flex-col">{children}</div>
		</div>
	);
}

/** Props injected into the dev chat view when it accepts shell callbacks. */
export type ShellMainChildProps = {
	onOpenSidebar?: () => void;
};
