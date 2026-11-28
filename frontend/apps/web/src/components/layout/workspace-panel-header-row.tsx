"use client";

import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { ConversationHeader } from "@/features/chat/components/conversation-header";
import { dmSidebarPeerMember } from "@/features/channel/lib/dm-sidebar-peer";
import { toChannelShellNavTarget } from "@/features/channel/lib/channel-shell-nav-target";
import { toQuickViewShellNavTarget } from "@/features/channel/lib/quick-view-shell-nav-target";
import type { QuickViewId } from "@/lib/app-routes";
import { useShellChrome } from "@/components/layout/shell-chrome-context";
import { useShellNavigation } from "@/components/layout/shell-navigation";
import { WorkspaceSidebarHeader } from "@/components/layout/workspace-sidebar-header";
import { WorkspaceGroupHeaderMenu } from "@/features/shell/components/workspace-group-header-menu";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import * as channelQuery from "@/services/channel/channel.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";
import { useMemo } from "react";
import { useSearchParams } from "next/navigation";

const QUICK_VIEW_LABEL_KEYS: Record<QuickViewId, string> = {
	mentions: "shell.sidebar.mentions",
	discover: "shell.sidebar.discover",
	drafts: "shell.sidebar.drafts",
	members: "group_settings.nav.members",
};

type WorkspacePanelHeaderRowProps = {
	onOpenSidebar?: () => void;
	groupId?: string;
	channelId?: string;
	quickView?: QuickViewId;
};

type WorkspaceSidebarHeaderCellProps = {
	groupId?: string;
	width: number;
	className?: string;
};

export function WorkspaceSidebarHeaderCell({
	groupId,
	width,
	className,
}: WorkspaceSidebarHeaderCellProps) {
	const myGroups = useMyGroupsQuery();
	const activeGroup = groupId ? myGroups.data?.find((g) => g.id === groupId) : undefined;
	const orderedGroupIds = (myGroups.data ?? []).map((g) => g.id);

	return (
		<div
			className={
				className ??
				"flex h-14 shrink-0 items-center border-r border-sidebar-border bg-sidebar lg:h-[46px]"
			}
			style={{ width }}
		>
			<div className="flex size-full items-center pl-4 pr-1.5">
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
	);
}

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
	const searchParams = useSearchParams();
	const mentionsPreviewChannelId =
		quickView === "mentions" ? searchParams.get("channelId")?.trim() : undefined;
	const effectiveChannelId = channelId ?? mentionsPreviewChannelId;
	const channels = useQuery({
		...channelQuery.groupChannelsQueryOpts(groupId ?? ""),
		enabled: Boolean(groupId && effectiveChannelId),
	});
	const { data: sessionUser } = useQuery(sessionQueryOpts());
	const membersQuery = useQuery({
		...groupMembersQueryOpts(groupId ?? "", ""),
		enabled: Boolean(groupId && effectiveChannelId),
	});
	const active = useMemo(() => {
		if (quickView && !mentionsPreviewChannelId) {
			return toQuickViewShellNavTarget(quickView, t(QUICK_VIEW_LABEL_KEYS[quickView]));
		}
		if (!groupId || !effectiveChannelId) {
			return mockActive;
		}
		const channel = channels.data?.channels.find((c) => c.id === effectiveChannelId);
		const peer =
			channel?.type === "dm"
				? dmSidebarPeerMember(channel, membersQuery.data?.members, sessionUser?.id)
				: undefined;
		return toChannelShellNavTarget(effectiveChannelId, channels.data?.channels, peer);
	}, [
		quickView,
		mentionsPreviewChannelId,
		t,
		groupId,
		effectiveChannelId,
		mockActive,
		channels.data?.channels,
		membersQuery.data?.members,
		sessionUser?.id,
	]);

	return (
		<div className="hidden shrink-0 border-b border-black lg:flex">
			<WorkspaceSidebarHeaderCell groupId={groupId} width={sidebarWidth} />
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
