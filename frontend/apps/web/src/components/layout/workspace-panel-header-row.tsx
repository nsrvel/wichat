"use client";

import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { ConversationHeader } from "@/features/chat/components/conversation-header";
import { toChannelShellNavTarget } from "@/features/channel/lib/channel-shell-nav-target";
import { toQuickViewShellNavTarget } from "@/features/channel/lib/quick-view-shell-nav-target";
import type { QuickViewId } from "@/lib/app-routes";
import { useShellChrome } from "@/components/layout/shell-chrome-context";
import { useShellNavigation } from "@/components/layout/shell-navigation";
import { WorkspaceSidebarHeader } from "@/components/layout/workspace-sidebar-header";
import { WorkspaceGroupHeaderMenu } from "@/features/shell/components/workspace-group-header-menu";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";
import * as channelQuery from "@/services/channel/channel.query";

const QUICK_VIEW_LABEL_KEYS: Record<QuickViewId, string> = {
	mentions: "shell.sidebar.mentions",
	discover: "shell.sidebar.discover",
	drafts: "shell.sidebar.drafts",
};

type WorkspacePanelHeaderRowProps = {
	onOpenSidebar?: () => void;
	groupId?: string;
	channelId?: string;
	quickView?: QuickViewId;
};

/** Sidebar + conversation headers — aligned strips, full-width divider like workspace chrome. */
export function WorkspacePanelHeaderRow({
	onOpenSidebar,
	groupId,
	channelId,
	quickView,
}: WorkspacePanelHeaderRowProps) {
	const { t } = useTranslation();
	const { sidebarWidth } = useShellChrome();
	const { active: mockActive } = useShellNavigation();
	const channels = useQuery({
		...channelQuery.groupChannelsQueryOpts(groupId ?? ""),
		enabled: Boolean(groupId && channelId),
	});
	const active = quickView
		? toQuickViewShellNavTarget(quickView, t(QUICK_VIEW_LABEL_KEYS[quickView]))
		: groupId && channelId
			? toChannelShellNavTarget(channelId, channels.data?.channels)
			: mockActive;
	const myGroups = useMyGroupsQuery();
	const activeGroup = groupId ? myGroups.data?.find((g) => g.id === groupId) : undefined;
	const orderedGroupIds = (myGroups.data ?? []).map((g) => g.id);

	return (
		<div className="hidden shrink-0 border-b border-black lg:flex">
			<div
				className="flex h-14 shrink-0 items-center border-r border-sidebar-border bg-sidebar lg:h-[46px]"
				style={{ width: sidebarWidth }}
			>
				<div className="flex h-full w-full items-center pl-4 pr-1.5">
					{activeGroup ? (
						<WorkspaceGroupHeaderMenu
							group={activeGroup}
							activeGroupId={groupId}
							orderedGroupIds={orderedGroupIds}
							className="min-w-0 flex-1"
						/>
					) : (
						<WorkspaceSidebarHeader
							title={myGroups.data?.[0]?.name ?? ""}
							className="min-w-0 flex-1"
						/>
					)}
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
