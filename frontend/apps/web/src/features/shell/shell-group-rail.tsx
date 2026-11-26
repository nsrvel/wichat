"use client";

import { useQueryClient } from "@tanstack/react-query";
import { Download, Plus, Store } from "lucide-react";
import { useChannelNavigation } from "@/features/channel/channel-navigation-provider";
import { useTranslation } from "react-i18next";
import { ICON_RAIL_WIDTH_PX } from "@/components/layout/shell-chrome-context";
import {
	shellRailAsideClass,
	shellRailAsideInnerClass,
	shellRailChromeState,
} from "@/components/layout/shell-rail/shell-rail-aside";
import { IconRailActionRow } from "@/components/layout/shell-rail/icon-rail-row";
import { IconRailTooltipTile } from "@/components/layout/shell-rail/icon-rail-tile-button";
import { useGroupOnboarding } from "@/features/group/group-onboarding-provider";
import { GroupRailAddGroupContextMenu } from "@/features/shell/components/group-rail-add-group-context-menu";
import { ShellGroupRailList } from "@/features/shell/components/shell-group-rail-list";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";
import { APP_DOWNLOADS_PATH } from "@/lib/app-routes";
import { readLastChannel } from "@/lib/last-channel-storage";
import { cn } from "@/lib/utils";
import * as channelQuery from "@/services/channel/channel.query";
import type { MyGroup } from "@/services/group/group.service";

export const SHELL_GROUP_RAIL_WIDTH_PX = ICON_RAIL_WIDTH_PX;

type ShellGroupRailProps = {
	open: boolean;
	activeGroupId?: string;
	activeChannelId?: string;
	className?: string;
};

export function ShellGroupRail({
	open,
	activeGroupId,
	activeChannelId,
	className,
}: ShellGroupRailProps) {
	const { t } = useTranslation();
	const queryClient = useQueryClient();
	const { navigateToChannel } = useChannelNavigation();
	const { openOnboarding } = useGroupOnboarding();
	const myGroups = useMyGroupsQuery();
	const groups = myGroups.data ?? [];

	const resolveChannelId = (groupId: string, defaultChannelId: string) => {
		const last = readLastChannel();
		if (last?.groupId === groupId) {
			return last.channelId;
		}
		if (activeGroupId === groupId && activeChannelId) {
			return activeChannelId;
		}
		return defaultChannelId;
	};

	const prefetchGroupChannels = (groupId: string) => {
		void queryClient.prefetchQuery(channelQuery.groupChannelsQueryOpts(groupId));
	};

	const navigateGroup = (group: MyGroup) => {
		if (group.id === activeGroupId) {
			return;
		}
		const channelId = resolveChannelId(group.id, group.default_channel_id);
		void queryClient.prefetchQuery(channelQuery.groupChannelsQueryOpts(group.id));
		navigateToChannel(group.id, channelId);
	};

	return (
		<aside
			className={cn(shellRailAsideClass(), className)}
			data-state={shellRailChromeState(open)}
			aria-label={t("shell.group_rail_label")}
			aria-hidden={!open}
		>
			<div className={shellRailAsideInnerClass()}>
				<ShellGroupRailList
					groups={groups}
					activeGroupId={activeGroupId}
					activeChannelId={activeChannelId}
					onNavigate={navigateGroup}
					onPrefetchGroup={prefetchGroupChannels}
				/>
			</div>
			<div className={shellRailAsideInnerClass()}>
				<IconRailActionRow>
					<GroupRailAddGroupContextMenu
						onJoin={() => openOnboarding("join")}
						onCreate={() => openOnboarding("create")}
					>
						<IconRailTooltipTile
							variant="muted"
							label={t("shell.add_group")}
							onClick={() => openOnboarding("pick")}
						>
							<Plus aria-hidden className="size-[18px]" strokeWidth={2} />
						</IconRailTooltipTile>
					</GroupRailAddGroupContextMenu>
				</IconRailActionRow>
				<IconRailActionRow>
					<IconRailTooltipTile variant="muted" label={t("shell.marketplace")}>
						<Store aria-hidden className="size-[18px]" strokeWidth={2} />
					</IconRailTooltipTile>
				</IconRailActionRow>
				<IconRailActionRow>
					<IconRailTooltipTile
						variant="muted"
						label={t("shell.download_apps")}
						onClick={() => window.open(APP_DOWNLOADS_PATH, "_blank", "noopener,noreferrer")}
					>
						<Download aria-hidden className="size-[18px]" strokeWidth={2} />
					</IconRailTooltipTile>
				</IconRailActionRow>
			</div>
		</aside>
	);
}
