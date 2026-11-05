"use client";

import { Download, Plus, Store } from "lucide-react";
import Image from "next/image";
import { useGroupOnboarding } from "@/features/group/group-onboarding-provider";
import { useRouter } from "next/navigation";
import { useTranslation } from "react-i18next";
import {
	shellRailAsideClass,
	shellRailAsideInnerClass,
	shellRailChromeState,
} from "@/components/layout/shell-rail/shell-rail-aside";
import { IconRailTooltipTile } from "@/components/layout/shell-rail/icon-rail-tile-button";
import {
	IconRailActionRow,
	IconRailWorkspaceRow,
} from "@/components/layout/shell-rail/icon-rail-row";
import { ICON_RAIL_WIDTH_PX } from "@/components/layout/shell-chrome-context";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";
import { channelPath } from "@/lib/app-routes";
import { readLastChannel } from "@/lib/last-channel-storage";
import { cn } from "@/lib/utils";
import { GroupRailTileContextMenu } from "@/features/shell/components/group-rail-tile-context-menu";
import { groupInitials } from "@/features/shell/lib/group-initials";
import { resolveGroupLogoUrl } from "@/features/shell/lib/group-logo-url";

export const SHELL_GROUP_RAIL_WIDTH_PX = ICON_RAIL_WIDTH_PX;

const railLogoTileClass =
	"relative overflow-hidden border-0 bg-transparent p-0 shadow-none hover:opacity-90";

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
	const router = useRouter();
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

	return (
		<aside
			className={cn(shellRailAsideClass(), className)}
			data-state={shellRailChromeState(open)}
			aria-label={t("shell.group_rail_label")}
			aria-hidden={!open}
		>
			<div className={shellRailAsideInnerClass()}>
				{groups.map((group) => {
					const active = group.id === activeGroupId;
					const logoSrc = resolveGroupLogoUrl(group.logo_url);
					return (
						<IconRailWorkspaceRow key={group.id} selected={active}>
							<GroupRailTileContextMenu group={group} activeGroupId={activeGroupId}>
								<IconRailTooltipTile
									variant="tile"
									className={logoSrc ? railLogoTileClass : undefined}
									label={group.name}
									aria-current={active ? "true" : undefined}
									onClick={() => {
										if (active) {
											return;
										}
										const channelId = resolveChannelId(group.id, group.default_channel_id);
										router.push(channelPath(group.id, channelId));
									}}
								>
									{logoSrc ? (
										<Image
											src={logoSrc}
											alt=""
											width={40}
											height={40}
											className="h-10 w-10 rounded-[13px] object-cover"
										/>
									) : (
										<span className="text-xs font-semibold">{groupInitials(group.name)}</span>
									)}
								</IconRailTooltipTile>
							</GroupRailTileContextMenu>
						</IconRailWorkspaceRow>
					);
				})}
			</div>
			<div className={shellRailAsideInnerClass()}>
				<IconRailActionRow>
					<IconRailTooltipTile
						variant="muted"
						label={t("shell.add_group")}
						onClick={() => openOnboarding()}
					>
						<Plus aria-hidden className="h-[18px] w-[18px]" strokeWidth={2} />
					</IconRailTooltipTile>
				</IconRailActionRow>
				<IconRailActionRow>
					<IconRailTooltipTile variant="muted" label={t("shell.marketplace")}>
						<Store aria-hidden className="h-[18px] w-[18px]" strokeWidth={2} />
					</IconRailTooltipTile>
				</IconRailActionRow>
				<IconRailActionRow>
					<IconRailTooltipTile variant="muted" label={t("shell.download_apps")}>
						<Download aria-hidden className="h-[18px] w-[18px]" strokeWidth={2} />
					</IconRailTooltipTile>
				</IconRailActionRow>
			</div>
		</aside>
	);
}
