"use client";

import { Hash } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslation } from "react-i18next";
import { SidebarListIconBox } from "@/components/layout/sidebar-list-icon";
import { WorkspaceSidebarHeader } from "@/components/layout/workspace-sidebar-header";
import {
	CHANNEL_SIDEBAR_MIN_WIDTH,
	useResizableSidebarWidth,
} from "@/hooks/use-resizable-sidebar-width";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";
import { channelPath } from "@/lib/app-routes";
import { cn } from "@/lib/utils";
import * as channelQuery from "@/services/channel/channel.query";
import { useQuery } from "@tanstack/react-query";

type ProductionChannelSidebarProps = {
	groupId: string;
	activeChannelId: string;
	width?: number;
	panelResizeHandle?: boolean;
};

export function ProductionChannelSidebar({
	groupId,
	activeChannelId,
	width: controlledWidth,
	panelResizeHandle = false,
}: ProductionChannelSidebarProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const myGroups = useMyGroupsQuery();
	const group = myGroups.data?.find((g) => g.id === groupId);
	const channels = useQuery(channelQuery.groupChannelsQueryOpts(groupId));

	const ownsResize = controlledWidth === undefined;
	const { width: internalWidth } = useResizableSidebarWidth(ownsResize);
	const width = controlledWidth ?? internalWidth;

	return (
		<div
			className="relative flex h-full shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground"
			style={{ width: controlledWidth ?? width }}
		>
			<div className="flex h-14 shrink-0 items-center border-b border-black lg:h-[46px]">
				<div className="flex h-full w-full items-center pl-4 pr-1.5">
					<WorkspaceSidebarHeader
						title={group?.name ?? t("shell.loading")}
						className="min-w-0 flex-1"
					/>
				</div>
			</div>
			<nav className="thin-scroll flex-1 overflow-y-auto px-3.5 py-2">
				<p className="mb-1 px-1.5 text-xs font-medium text-muted-foreground">
					{t("shell.channel_list_label")}
				</p>
				<ul className="flex flex-col gap-0.5">
					{(channels.data?.channels ?? []).map((ch) => {
						const selected = ch.id === activeChannelId;
						const label = ch.name === "general" ? `#${ch.name}` : ch.name;
						return (
							<li key={ch.id}>
								<button
									type="button"
									className={cn(
										"flex w-full min-h-[32px] items-center gap-2 rounded-md px-2 text-left text-sm transition-colors hover:bg-sidebar-accent",
										selected && "bg-sidebar-accent font-semibold text-sidebar-foreground",
									)}
									aria-current={selected ? "true" : undefined}
									onClick={() => router.push(channelPath(groupId, ch.id))}
								>
									<SidebarListIconBox>
										<Hash className="h-[11px] w-[11px] text-[#8e8e93]" strokeWidth={2.2} />
									</SidebarListIconBox>
									<span className="truncate">{label}</span>
								</button>
							</li>
						);
					})}
				</ul>
			</nav>
			{!panelResizeHandle ? null : (
				<div className="pointer-events-none absolute inset-y-0 right-0 w-px" aria-hidden />
			)}
		</div>
	);
}

export { CHANNEL_SIDEBAR_MIN_WIDTH };
