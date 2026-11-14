"use client";

import { AtSign, Compass, FileText } from "lucide-react";
import { useEffect, useState } from "react";
import { SidebarCategoryRow } from "@/components/layout/channel-sidebar/sidebar-category-row";
import { SidebarChannelRow } from "@/components/layout/channel-sidebar/sidebar-channel-row";
import { SidebarFavoriteDmRow } from "@/components/layout/channel-sidebar/sidebar-favorite-dm-row";
import { SidebarNavItem } from "@/components/layout/channel-sidebar/sidebar-nav-item";
import {
	SIDEBAR_INSET_X,
	sidebarNestedListClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { useShellChrome } from "@/components/layout/shell-chrome-context";
import { ChannelListIcon } from "@/components/layout/channel-list-icon";
import {
	channelNavBellIcon,
	channelNavHashIcon,
	channelNavInfraIcon,
	channelNavWichatIcon,
	channelNavWicloudIcon,
	channelNavWinobarIcon,
	channelNavWitaskIcon,
} from "@/components/layout/channel-nav-icon";
import {
	ChannelGlyphSlot,
	SidebarListAvatar,
	sidebarNavGlyphIconClass,
} from "@/components/layout/sidebar-list-icon";
import { useShellNavigation, type ShellNavKind } from "@/components/layout/shell-navigation";
import { WorkspaceSidebarFilter } from "@/components/layout/workspace-sidebar-filter";
import { WorkspaceSidebarHeader } from "@/components/layout/workspace-sidebar-header";
import {
	CHANNEL_SIDEBAR_MIN_WIDTH,
	useResizableSidebarWidth,
} from "@/hooks/use-resizable-sidebar-width";
import { cn } from "@/lib/utils";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";

type ChannelSidebarProps = {
	/** Desktop: drag to resize. Mobile drawer should pass `false`. */
	resizable?: boolean;
	/** `list` — channel list; header stays in WorkspacePanel. */
	layout?: "full" | "list";
	/** Controlled width from `WorkspacePanel` (full-height resize handle). */
	width?: number;
	/** Resize handle rendered on workspace panel, not on this column. */
	panelResizeHandle?: boolean;
};

export function ChannelSidebar({
	resizable = true,
	layout = "full",
	width: controlledWidth,
	panelResizeHandle = false,
}: ChannelSidebarProps) {
	const myGroups = useMyGroupsQuery();
	const groupName = myGroups.data?.[0]?.name ?? "";
	const ownsResize = resizable && controlledWidth === undefined;
	const { width: internalWidth, onResizePointerDown } = useResizableSidebarWidth(ownsResize);
	const width = controlledWidth ?? internalWidth;
	const { setSidebarWidth } = useShellChrome();

	useEffect(() => {
		if (controlledWidth !== undefined) return;
		setSidebarWidth(width);
	}, [controlledWidth, width, setSidebarWidth]);
	const [quickViewsOpen, setQuickViewsOpen] = useState(true);
	const [favoritesOpen, setFavoritesOpen] = useState(true);
	const [channelsOpen, setChannelsOpen] = useState(true);
	const [chatsOpen, setChatsOpen] = useState(true);
	const [activeMenuId, setActiveMenuId] = useState<string | null>(null);
	const { navigate, isActive } = useShellNavigation();

	const selectNav = (id: string, label: string, kind: ShellNavKind) => () => {
		navigate({ id, label, kind });
	};

	return (
		<div
			className="relative flex h-full shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground"
			style={{ width: resizable ? width : CHANNEL_SIDEBAR_MIN_WIDTH }}
		>
			{layout === "full" ? (
				<div className="flex h-14 shrink-0 items-center border-b border-black lg:h-[46px]">
					<div className="flex size-full items-center pl-4 pr-1.5">
						<WorkspaceSidebarHeader title={groupName} className="min-w-0 flex-1" />
					</div>
				</div>
			) : (
				<div className="flex shrink-0 items-center gap-2 border-b border-black py-1 pl-4 pr-1.5">
					<WorkspaceSidebarFilter className="min-w-0 flex-1 border-0 p-0" />
				</div>
			)}
			<div className={cn("flex min-h-0 flex-1 flex-col", SIDEBAR_INSET_X)}>
				{layout === "full" ? <WorkspaceSidebarFilter /> : null}
				<div className="thin-scroll flex-1 overflow-y-auto">
					<div className="group/cat">
						<SidebarCategoryRow
							title="Quick views"
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
									icon={<AtSign className={sidebarNavGlyphIconClass} strokeWidth={2} />}
									label="Mentions"
								/>
								<SidebarNavItem
									menuId="qv-discover"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<Compass className={sidebarNavGlyphIconClass} strokeWidth={2} />}
									label="Discover"
								/>
								<SidebarNavItem
									menuId="qv-drafts"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<FileText className={sidebarNavGlyphIconClass} strokeWidth={2} />}
									label="Drafts"
								/>
							</nav>
						) : null}
					</div>
					<div className="group/cat mt-2">
						<SidebarCategoryRow
							title="Favorites"
							open={favoritesOpen}
							onToggle={() => setFavoritesOpen((o) => !o)}
							menuId="cat-favorites"
							activeMenuId={activeMenuId}
							onActiveMenuIdChange={setActiveMenuId}
						/>
						{favoritesOpen ? (
							<div className={sidebarNestedListClass}>
								<SidebarFavoriteDmRow
									menuId="fav-self-dm"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									avatarSrc="/avatars/demo/marvell.png"
									label="ⱮąɾѵҽӀӀ"
									selected={isActive("fav-self-dm")}
									onSelect={selectNav("fav-self-dm", "ⱮąɾѵҽӀӀ", "dm")}
								/>
								<SidebarChannelRow
									menuId="fav-announcements"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<ChannelListIcon type="broadcast" channelName="Announcements" />}
									label="Announcements"
									selected={isActive("fav-announcements")}
									onSelect={selectNav("fav-announcements", "Announcements", "channel")}
								/>
								<SidebarChannelRow
									menuId="fav-deployment-production"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<ChannelGlyphSlot>{channelNavBellIcon}</ChannelGlyphSlot>}
									label="Deployment Production"
									selected={isActive("fav-deployment-production")}
									onSelect={selectNav(
										"fav-deployment-production",
										"Deployment Production",
										"channel",
									)}
								/>
								<SidebarChannelRow
									menuId="fav-deployment-staging"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<ChannelGlyphSlot>{channelNavBellIcon}</ChannelGlyphSlot>}
									label="Deployment Staging"
									selected={isActive("fav-deployment-staging")}
									onSelect={selectNav("fav-deployment-staging", "Deployment Staging", "channel")}
								/>
								<SidebarChannelRow
									menuId="fav-release-notes"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<ChannelGlyphSlot>{channelNavHashIcon}</ChannelGlyphSlot>}
									label="Release Notes"
									selected={isActive("fav-release-notes")}
									onSelect={selectNav("fav-release-notes", "Release Notes", "channel")}
								/>
							</div>
						) : null}
					</div>
					<div className="group/cat mt-2">
						<SidebarCategoryRow
							title="Channels"
							open={channelsOpen}
							onToggle={() => setChannelsOpen((o) => !o)}
							menuId="cat-channels"
							activeMenuId={activeMenuId}
							onActiveMenuIdChange={setActiveMenuId}
						/>
						{channelsOpen ? (
							<div className={sidebarNestedListClass}>
								<SidebarChannelRow
									menuId="ch-infra"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<ChannelGlyphSlot>{channelNavInfraIcon}</ChannelGlyphSlot>}
									label="Infra"
									selected={isActive("ch-infra")}
									onSelect={selectNav("ch-infra", "Infra", "channel")}
								/>
								<SidebarChannelRow
									menuId="ch-wichat"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<ChannelGlyphSlot>{channelNavWichatIcon}</ChannelGlyphSlot>}
									label="WiChat"
									selected={isActive("ch-wichat")}
									onSelect={selectNav("ch-wichat", "WiChat", "channel")}
								/>
								<SidebarChannelRow
									menuId="ch-wicloud"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<ChannelGlyphSlot>{channelNavWicloudIcon}</ChannelGlyphSlot>}
									label="WiCloud"
									selected={isActive("ch-wicloud")}
									onSelect={selectNav("ch-wicloud", "WiCloud", "channel")}
								/>
								<SidebarChannelRow
									menuId="ch-witask"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<ChannelGlyphSlot>{channelNavWitaskIcon}</ChannelGlyphSlot>}
									label="WiTask"
									selected={isActive("ch-witask")}
									onSelect={selectNav("ch-witask", "WiTask", "channel")}
								/>
								<SidebarChannelRow
									menuId="ch-winobar"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<ChannelGlyphSlot>{channelNavWinobarIcon}</ChannelGlyphSlot>}
									label="WiNobar"
									selected={isActive("ch-winobar")}
									onSelect={selectNav("ch-winobar", "WiNobar", "channel")}
								/>
							</div>
						) : null}
					</div>
					<div className="group/cat mt-2">
						<SidebarCategoryRow
							title="Chats"
							open={chatsOpen}
							onToggle={() => setChatsOpen((o) => !o)}
							menuId="cat-chats"
							activeMenuId={activeMenuId}
							onActiveMenuIdChange={setActiveMenuId}
						/>
						{chatsOpen ? (
							<div className={sidebarNestedListClass}>
								<SidebarChannelRow
									menuId="chat-dm-koh-yadi"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<SidebarListAvatar src="/avatars/demo/koh-yadi.jpg" alt="Koh Yadi" />}
									label="Koh Yadi"
									selected={isActive("chat-dm-koh-yadi")}
									onSelect={selectNav("chat-dm-koh-yadi", "Koh Yadi", "dm")}
								/>
								<SidebarChannelRow
									menuId="chat-dm-koh-yanto"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<SidebarListAvatar src="/avatars/demo/koh-yanto.jpg" alt="Koh Yanto" />}
									label="Koh Yanto"
									selected={isActive("chat-dm-koh-yanto")}
									onSelect={selectNav("chat-dm-koh-yanto", "Koh Yanto", "dm")}
								/>
							</div>
						) : null}
					</div>
				</div>
			</div>
			{resizable && !panelResizeHandle ? (
				<div
					role="separator"
					aria-orientation="vertical"
					aria-label="Resize sidebar"
					className="absolute top-0 right-0 z-10 h-full w-1.5 translate-x-1/2 cursor-col-resize touch-none hover:bg-primary/15 active:bg-primary/25"
					onPointerDown={onResizePointerDown}
				/>
			) : null}
		</div>
	);
}
