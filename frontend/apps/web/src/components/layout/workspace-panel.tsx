"use client";

import type { ReactNode } from "react";
import { useEffect } from "react";
import { ChannelSidebar } from "@/components/layout/channel-sidebar";
import { useShellChrome } from "@/components/layout/shell-chrome-context";
import { WorkspacePanelHeaderRow } from "@/components/layout/workspace-panel-header-row";
import { sidebarMinWidthForUserDock } from "@/components/layout/shell-user-dock";
import { WorkspaceSidebarResizeHandle } from "@/components/layout/workspace-sidebar-resize-handle";
import {
	CHANNEL_SIDEBAR_MIN_WIDTH,
	useResizableSidebarWidth,
} from "@/hooks/use-resizable-sidebar-width";

type WorkspacePanelProps = {
	children: ReactNode;
	onOpenSidebar?: () => void;
	/** Production channel list; defaults to mock `ChannelSidebar` for `/dev/chat-ui`. */
	sidebar?: ReactNode;
};

export function WorkspacePanel({ children, onOpenSidebar, sidebar }: WorkspacePanelProps) {
	const { railOpen, setSidebarWidth } = useShellChrome();
	const sidebarMinWidth = sidebarMinWidthForUserDock(railOpen, CHANNEL_SIDEBAR_MIN_WIDTH);
	const { width, onResizePointerDown } = useResizableSidebarWidth(true, sidebarMinWidth);

	useEffect(() => {
		setSidebarWidth(width);
	}, [width, setSidebarWidth]);

	return (
		<div className="relative flex h-full min-h-0 min-w-0 flex-col overflow-hidden">
			<WorkspacePanelHeaderRow onOpenSidebar={onOpenSidebar} />
			<div className="flex min-h-0 min-w-0 flex-1">
				<div className="hidden h-full min-h-0 shrink-0 lg:block" style={{ width }}>
					{sidebar ?? <ChannelSidebar layout="list" width={width} panelResizeHandle />}
				</div>
				<div className="flex min-h-0 min-w-0 flex-1 flex-col bg-chat-surface">{children}</div>
			</div>
			<WorkspaceSidebarResizeHandle
				sidebarWidth={width}
				className="hidden lg:block"
				onPointerDown={onResizePointerDown}
			/>
		</div>
	);
}
