"use client";

import { cloneElement, isValidElement, useEffect, type ReactElement, type ReactNode } from "react";
import { ChannelSidebar } from "@/components/layout/channel-sidebar";
import { useShellChrome } from "@/components/layout/shell-chrome-context";
import { ShellUserDock } from "@/components/layout/shell-user-dock";
import {
	WorkspacePanelHeaderRow,
	WorkspaceSidebarHeaderCell,
} from "@/components/layout/workspace-panel-header-row";
import { WorkspaceSidebarResizeHandle } from "@/components/layout/workspace-sidebar-resize-handle";
import {
	CHANNEL_SIDEBAR_MIN_WIDTH,
	useResizableSidebarWidth,
} from "@/hooks/use-resizable-sidebar-width";
import type { QuickViewId } from "@/lib/app-routes";

type WorkspacePanelProps = {
	children: ReactNode;
	onOpenSidebar?: () => void;
	groupId?: string;
	channelId?: string;
	quickView?: QuickViewId;
	/** Production channel list; defaults to mock `ChannelSidebar` for `/dev/chat-ui`. */
	sidebar?: ReactNode;
};

type SidebarWidthProps = {
	width?: number;
	panelResizeHandle?: boolean;
};

function sidebarWithPanelWidth(sidebar: ReactNode, width: number): ReactNode {
	if (!isValidElement(sidebar)) {
		return sidebar;
	}
	return cloneElement(sidebar as ReactElement<SidebarWidthProps>, {
		width,
		panelResizeHandle: false,
	});
}

export function WorkspacePanel({
	children,
	onOpenSidebar,
	groupId,
	channelId,
	quickView,
	sidebar,
}: WorkspacePanelProps) {
	const { setSidebarWidth } = useShellChrome();
	const { width, onResizePointerDown } = useResizableSidebarWidth(true, CHANNEL_SIDEBAR_MIN_WIDTH);

	useEffect(() => {
		setSidebarWidth(width);
	}, [width, setSidebarWidth]);

	const sidebarNode =
		sidebar != null ? (
			sidebarWithPanelWidth(sidebar, width)
		) : (
			<ChannelSidebar layout="list" width={width} panelResizeHandle />
		);

	const isMentionsSplit = quickView === "mentions";

	return (
		<div className="relative flex h-full min-h-0 min-w-0 flex-col overflow-hidden">
			{!isMentionsSplit ? (
				<WorkspacePanelHeaderRow
					onOpenSidebar={onOpenSidebar}
					groupId={groupId}
					channelId={channelId}
					quickView={quickView}
				/>
			) : null}
			<div className="flex min-h-0 min-w-0 flex-1">
				<div className="hidden h-full min-h-0 shrink-0 lg:flex lg:flex-col" style={{ width }}>
					{isMentionsSplit ? (
						<WorkspaceSidebarHeaderCell
							groupId={groupId}
							width={width}
							className="flex h-[46px] shrink-0 items-center border-b border-black border-r border-sidebar-border bg-sidebar"
						/>
					) : null}
					<div className="min-h-0 flex-1">{sidebarNode}</div>
				</div>
				<div className="flex min-h-0 min-w-0 flex-1 flex-col bg-chat-surface">{children}</div>
			</div>
			<ShellUserDock />
			<WorkspaceSidebarResizeHandle
				sidebarWidth={width}
				className="hidden lg:block"
				onPointerDown={onResizePointerDown}
			/>
		</div>
	);
}
