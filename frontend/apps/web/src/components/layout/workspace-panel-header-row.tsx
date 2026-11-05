"use client";

import { ConversationHeader } from "@/features/chat/components/conversation-header";
import { useShellChrome } from "@/components/layout/shell-chrome-context";
import { useShellNavigation } from "@/components/layout/shell-navigation";
import { WorkspaceSidebarHeader } from "@/components/layout/workspace-sidebar-header";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";
type WorkspacePanelHeaderRowProps = {
	onOpenSidebar?: () => void;
};

/** Sidebar + conversation headers — aligned strips, full-width divider like workspace chrome. */
export function WorkspacePanelHeaderRow({ onOpenSidebar }: WorkspacePanelHeaderRowProps) {
	const { sidebarWidth } = useShellChrome();
	const { active } = useShellNavigation();
	const myGroups = useMyGroupsQuery();
	const groupName = myGroups.data?.[0]?.name ?? "";

	return (
		<div className="hidden shrink-0 border-b border-black lg:flex">
			<div
				className="flex h-14 shrink-0 items-center border-r border-sidebar-border bg-sidebar lg:h-[46px]"
				style={{ width: sidebarWidth }}
			>
				<div className="flex h-full w-full items-center pl-4 pr-1.5">
					<WorkspaceSidebarHeader title={groupName} className="min-w-0 flex-1" />
				</div>
			</div>
			<div className="min-w-0 flex-1">
				<ConversationHeader
					active={active}
					onOpenSidebar={onOpenSidebar}
					showBottomDivider={false}
				/>
			</div>
		</div>
	);
}
