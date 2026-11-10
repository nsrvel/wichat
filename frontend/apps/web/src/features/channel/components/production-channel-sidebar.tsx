"use client";

import { AtSign, Bell, Compass, FileText, Hash } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { SidebarCategoryEmptyHint } from "@/components/layout/channel-sidebar/sidebar-category-empty-hint";
import { SidebarCategoryRow } from "@/components/layout/channel-sidebar/sidebar-category-row";
import { SidebarChannelRow } from "@/components/layout/channel-sidebar/sidebar-channel-row";
import { SidebarNavItem } from "@/components/layout/channel-sidebar/sidebar-nav-item";
import {
	SIDEBAR_INSET_X,
	sidebarNestedListClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import {
	ChannelGlyphSlot,
	channelListGlyphIconClass,
	sidebarNavGlyphIconClass,
} from "@/components/layout/sidebar-list-icon";
import {
	CHANNEL_SIDEBAR_MIN_WIDTH,
	useResizableSidebarWidth,
} from "@/hooks/use-resizable-sidebar-width";
import { channelPath, quickViewPath, type QuickViewId } from "@/lib/app-routes";
import { cn } from "@/lib/utils";
import { partitionChannelsBySidebarSection } from "@/features/channel/lib/partition-group-channels";
import * as channelQuery from "@/services/channel/channel.query";
import { useQuery } from "@tanstack/react-query";

type ProductionChannelSidebarProps = {
	groupId: string;
	activeChannelId: string;
	activeQuickView?: QuickViewId;
	width?: number;
	panelResizeHandle?: boolean;
};

export function ProductionChannelSidebar({
	groupId,
	activeChannelId,
	activeQuickView,
	width: controlledWidth,
	panelResizeHandle = false,
}: ProductionChannelSidebarProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const channels = useQuery(channelQuery.groupChannelsQueryOpts(groupId));

	const embeddedInPanel = controlledWidth !== undefined;
	const { width: internalWidth } = useResizableSidebarWidth(!embeddedInPanel);
	const width = controlledWidth ?? internalWidth;

	const [quickViewsOpen, setQuickViewsOpen] = useState(true);
	const [favoritesOpen, setFavoritesOpen] = useState(true);
	const [channelsOpen, setChannelsOpen] = useState(true);
	const [chatsOpen, setChatsOpen] = useState(true);
	const [activeMenuId, setActiveMenuId] = useState<string | null>(null);

	const channelList = channels.data?.channels ?? [];
	const { favorites: favoriteChannels, listChannels } =
		partitionChannelsBySidebarSection(channelList);

	return (
		<div
			className={cn(
				"relative flex h-full flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground",
				embeddedInPanel ? "min-w-0 w-full" : "shrink-0",
			)}
			style={embeddedInPanel ? undefined : { width }}
		>
			<div className={cn("flex min-h-0 flex-1 flex-col pt-1", SIDEBAR_INSET_X)}>
				<div className="thin-scroll flex-1 overflow-y-auto">
					<div className="group/cat">
						<SidebarCategoryRow
							title={t("shell.sidebar.quick_views")}
							open={quickViewsOpen}
							onToggle={() => setQuickViewsOpen((o) => !o)}
							menuId="cat-quick-views"
							activeMenuId={activeMenuId}
							onActiveMenuIdChange={setActiveMenuId}
						/>
						{quickViewsOpen ? (
							<nav className={sidebarNestedListClass}>
								<SidebarNavItem
									menuId="qv-mentions"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									selected={activeQuickView === "mentions"}
									onSelect={() => router.push(quickViewPath(groupId, "mentions"))}
									icon={<AtSign className={sidebarNavGlyphIconClass} strokeWidth={2} />}
									label={t("shell.sidebar.mentions")}
								/>
								<SidebarNavItem
									menuId="qv-discover"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									selected={activeQuickView === "discover"}
									onSelect={() => router.push(quickViewPath(groupId, "discover"))}
									icon={<Compass className={sidebarNavGlyphIconClass} strokeWidth={2} />}
									label={t("shell.sidebar.discover")}
								/>
								<SidebarNavItem
									menuId="qv-drafts"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									selected={activeQuickView === "drafts"}
									onSelect={() => router.push(quickViewPath(groupId, "drafts"))}
									icon={<FileText className={sidebarNavGlyphIconClass} strokeWidth={2} />}
									label={t("shell.sidebar.drafts")}
								/>
							</nav>
						) : null}
					</div>
					<div className="group/cat mt-2">
						<SidebarCategoryRow
							title={t("shell.sidebar.favorites")}
							open={favoritesOpen}
							onToggle={() => setFavoritesOpen((o) => !o)}
							menuId="cat-favorites"
							activeMenuId={activeMenuId}
							onActiveMenuIdChange={setActiveMenuId}
						/>
						{favoritesOpen ? (
							favoriteChannels.length > 0 ? (
								<div className={sidebarNestedListClass}>
									{favoriteChannels.map((ch) => (
										<SidebarChannelRow
											key={ch.id}
											menuId={`ch-${ch.id}`}
											activeMenuId={activeMenuId}
											onActiveMenuIdChange={setActiveMenuId}
											icon={
												<ChannelGlyphSlot tileTone="announcement">
													<Bell className={channelListGlyphIconClass} strokeWidth={2} />
												</ChannelGlyphSlot>
											}
											label={ch.name}
											selected={!activeQuickView && ch.id === activeChannelId}
											onSelect={() => router.push(channelPath(groupId, ch.id))}
										/>
									))}
								</div>
							) : (
								<SidebarCategoryEmptyHint>{t("shell.sidebar.empty_favorites")}</SidebarCategoryEmptyHint>
							)
						) : null}
					</div>
					<div className="group/cat mt-2">
						<SidebarCategoryRow
							title={t("shell.sidebar.channels")}
							open={channelsOpen}
							onToggle={() => setChannelsOpen((o) => !o)}
							menuId="cat-channels"
							activeMenuId={activeMenuId}
							onActiveMenuIdChange={setActiveMenuId}
						/>
						{channelsOpen ? (
							listChannels.length > 0 ? (
								<div className={sidebarNestedListClass}>
									{listChannels.map((ch) => (
										<SidebarChannelRow
											key={ch.id}
											menuId={`ch-${ch.id}`}
											activeMenuId={activeMenuId}
											onActiveMenuIdChange={setActiveMenuId}
											icon={
												<ChannelGlyphSlot tileTone="default">
													<Hash className={channelListGlyphIconClass} strokeWidth={2} />
												</ChannelGlyphSlot>
											}
											label={ch.name}
											selected={!activeQuickView && ch.id === activeChannelId}
											onSelect={() => router.push(channelPath(groupId, ch.id))}
										/>
									))}
								</div>
							) : channels.isPending ? null : (
								<SidebarCategoryEmptyHint>{t("shell.sidebar.empty_channels")}</SidebarCategoryEmptyHint>
							)
						) : null}
					</div>
					<div className="group/cat mt-2">
						<SidebarCategoryRow
							title={t("shell.sidebar.chats")}
							open={chatsOpen}
							onToggle={() => setChatsOpen((o) => !o)}
							menuId="cat-chats"
							activeMenuId={activeMenuId}
							onActiveMenuIdChange={setActiveMenuId}
						/>
						{chatsOpen ? (
							<SidebarCategoryEmptyHint>{t("shell.sidebar.empty_chats")}</SidebarCategoryEmptyHint>
						) : null}
					</div>
				</div>
			</div>
			{!panelResizeHandle ? null : (
				<div className="pointer-events-none absolute inset-y-0 right-0 w-px" aria-hidden />
			)}
		</div>
	);
}

export { CHANNEL_SIDEBAR_MIN_WIDTH };
