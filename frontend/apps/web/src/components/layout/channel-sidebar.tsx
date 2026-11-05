"use client";

import { AtSign, ChevronDown, Compass, FileText } from "lucide-react";
import { useEffect, useState, type ReactNode } from "react";
import { useShellChrome, useShellChromeDockInset } from "@/components/layout/shell-chrome-context";
import {
	channelNavBellIcon,
	channelNavHashIcon,
	channelNavInfraIcon,
	channelNavWichatIcon,
	channelNavWicloudIcon,
	channelNavWinobarIcon,
	channelNavWitaskIcon,
} from "@/components/layout/channel-nav-icon";
import { SidebarListAvatar, SidebarListIconBox } from "@/components/layout/sidebar-list-icon";
import { useShellNavigation, type ShellNavKind } from "@/components/layout/shell-navigation";
import { SidebarMoreMenu } from "@/components/layout/sidebar-more-menu";
import { WorkspaceSidebarFilter } from "@/components/layout/workspace-sidebar-filter";
import { WorkspaceSidebarHeader } from "@/components/layout/workspace-sidebar-header";
import {
	CHANNEL_SIDEBAR_MIN_WIDTH,
	useResizableSidebarWidth,
} from "@/hooks/use-resizable-sidebar-width";
import { cn } from "@/lib/utils";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";

/** Horizontal inset for header + nav. */
const SIDEBAR_INSET_X = "px-3.5";

const SIDEBAR_ROW_LG_H = "lg:min-h-[32px] lg:h-[32px]";

const sidebarRowLabelClass = (unread?: boolean) =>
	unread ? "font-semibold text-sidebar-foreground" : "font-medium text-muted-foreground";

const sidebarRowShellClass = cn(
	"group/row relative flex min-h-[30px] items-center rounded-md transition-colors hover:bg-sidebar-accent",
	SIDEBAR_ROW_LG_H,
);

const sidebarRowMoreRailClass = (menuOpen: boolean) =>
	cn(
		"absolute top-0 right-0 z-[1] flex h-full w-8 items-center justify-end pr-0.5 opacity-0 transition-opacity group-hover/row:opacity-100",
		menuOpen && "opacity-100",
	);

const sidebarRowInnerButtonClass = cn(
	"flex h-11 w-full min-w-0 items-center gap-3 rounded-md px-2.5 lg:gap-2 lg:px-1.5",
	SIDEBAR_ROW_LG_H,
);

/** Category header: same row height, chevron flush left (items indent via nested list). */
const sidebarCategoryInnerButtonClass = cn(
	"flex h-11 w-full min-w-0 flex-1 items-center gap-1 rounded-md pl-0.5 pr-2.5",
	SIDEBAR_ROW_LG_H,
);

const sidebarRowTextClass = "truncate text-sm lg:text-sm";

const sidebarNestedListClass = "flex flex-col gap-0 pl-3 lg:pl-2.5";

function SidebarCategoryRow({
	title,
	open,
	onToggle,
	menuId,
	activeMenuId,
	onActiveMenuIdChange,
}: {
	title: string;
	open: boolean;
	onToggle: () => void;
	menuId: string;
	activeMenuId: string | null;
	onActiveMenuIdChange: (id: string | null) => void;
}) {
	const menuOpen = activeMenuId === menuId;

	return (
		<div className={sidebarRowShellClass}>
			<button
				type="button"
				className={cn(
					sidebarCategoryInnerButtonClass,
					"min-w-0 flex-1 text-left",
					open && "text-foreground",
				)}
				aria-expanded={open}
				onClick={onToggle}
			>
				<span
					className={cn(
						"flex h-5 w-5 shrink-0 items-center justify-center text-muted-foreground",
						open && "text-foreground",
					)}
				>
					<ChevronDown
						className={cn("h-3.5 w-3.5 -rotate-90 transition-transform", open && "rotate-0")}
						strokeWidth={2.2}
					/>
				</span>
				<span className="min-w-0 flex-1 truncate text-xs font-medium text-muted-foreground lg:text-[13px] group-hover/row:font-semibold">
					{title}
				</span>
			</button>
			<div className={cn(sidebarRowMoreRailClass(menuOpen), "group-hover/cat:opacity-100")}>
				<SidebarMoreMenu
					id={menuId}
					activeId={activeMenuId}
					onActiveIdChange={onActiveMenuIdChange}
				/>
			</div>
		</div>
	);
}

function SidebarRowShell({
	menuId,
	activeMenuId,
	onActiveMenuIdChange,
	unread,
	selected,
	rowClassName,
	children,
}: {
	menuId: string;
	activeMenuId: string | null;
	onActiveMenuIdChange: (id: string | null) => void;
	unread?: boolean;
	selected?: boolean;
	rowClassName?: string;
	children: ReactNode;
}) {
	const menuOpen = activeMenuId === menuId;

	return (
		<div className={cn(sidebarRowShellClass, selected && "bg-sidebar-row-selected", rowClassName)}>
			<div className="min-w-0 flex-1">{children}</div>
			{unread ? (
				<span
					className={cn(
						"pointer-events-none absolute top-1/2 right-2 z-[2] h-1.5 w-1.5 -translate-y-1/2 rounded-full bg-primary transition-opacity",
						menuOpen ? "opacity-0" : "group-hover/row:opacity-0",
					)}
					aria-hidden
				/>
			) : null}
			<div className={sidebarRowMoreRailClass(menuOpen)}>
				<SidebarMoreMenu
					id={menuId}
					activeId={activeMenuId}
					onActiveIdChange={onActiveMenuIdChange}
				/>
			</div>
		</div>
	);
}

function FavoriteDmItem({
	avatarSrc,
	label,
	unread,
	selected,
	menuId,
	activeMenuId,
	onActiveMenuIdChange,
	onSelect,
}: {
	avatarSrc: string;
	label: string;
	unread?: boolean;
	selected?: boolean;
	menuId: string;
	activeMenuId: string | null;
	onActiveMenuIdChange: (id: string | null) => void;
	onSelect: () => void;
}) {
	return (
		<SidebarRowShell
			menuId={menuId}
			activeMenuId={activeMenuId}
			onActiveMenuIdChange={onActiveMenuIdChange}
			unread={unread}
			selected={selected}
		>
			<button
				type="button"
				className={cn(
					sidebarRowInnerButtonClass,
					selected ? "font-semibold text-sidebar-foreground" : sidebarRowLabelClass(unread),
				)}
				onClick={onSelect}
			>
				<SidebarListAvatar src={avatarSrc} />
				<span
					className={cn(
						sidebarRowTextClass,
						"min-w-0 flex-1 truncate text-left",
						sidebarRowLabelClass(unread),
					)}
				>
					{label}
				</span>
			</button>
		</SidebarRowShell>
	);
}

function NavItem({
	icon,
	label,
	badge,
	unread,
	menuId,
	activeMenuId,
	onActiveMenuIdChange,
}: {
	icon: ReactNode;
	label: string;
	badge?: boolean;
	unread?: boolean;
	menuId: string;
	activeMenuId: string | null;
	onActiveMenuIdChange: (id: string | null) => void;
}) {
	return (
		<SidebarRowShell
			menuId={menuId}
			activeMenuId={activeMenuId}
			onActiveMenuIdChange={onActiveMenuIdChange}
			unread={unread}
		>
			<button type="button" className={sidebarRowInnerButtonClass}>
				<SidebarListIconBox className="text-[#6b6b70]">{icon}</SidebarListIconBox>
				<span className={cn(sidebarRowTextClass, "min-w-0 flex-1 truncate text-left font-medium")}>
					{label}
				</span>
				{badge ? (
					<span className="rounded bg-[#F2D9F6] px-1 py-px text-[9px] font-semibold tracking-wide text-[#A855C8]">
						NEW
					</span>
				) : null}
			</button>
		</SidebarRowShell>
	);
}

function ChannelItem({
	icon,
	label,
	unread,
	selected,
	menuId,
	activeMenuId,
	onActiveMenuIdChange,
	onSelect,
}: {
	icon: ReactNode;
	label: string;
	unread?: boolean;
	selected?: boolean;
	menuId: string;
	activeMenuId: string | null;
	onActiveMenuIdChange: (id: string | null) => void;
	onSelect: () => void;
}) {
	return (
		<SidebarRowShell
			menuId={menuId}
			activeMenuId={activeMenuId}
			onActiveMenuIdChange={onActiveMenuIdChange}
			unread={unread}
			selected={selected}
		>
			<button
				type="button"
				className={cn(
					sidebarRowInnerButtonClass,
					sidebarRowTextClass,
					selected ? "font-semibold text-sidebar-foreground" : sidebarRowLabelClass(unread),
				)}
				onClick={onSelect}
			>
				{icon}
				<span className="min-w-0 flex-1 truncate text-left">{label}</span>
			</button>
		</SidebarRowShell>
	);
}

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
	const dockInset = useShellChromeDockInset();

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
					<div className="flex h-full w-full items-center pl-4 pr-1.5">
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
				<div className="thin-scroll flex-1 overflow-y-auto" style={{ paddingBottom: dockInset }}>
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
								<NavItem
									menuId="qv-mentions"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<AtSign className="h-[11px] w-[11px]" strokeWidth={2.2} />}
									label="Mentions"
								/>
								<NavItem
									menuId="qv-discover"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<Compass className="h-[11px] w-[11px]" strokeWidth={2.2} />}
									label="Discover"
								/>
								<NavItem
									menuId="qv-drafts"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<FileText className="h-[11px] w-[11px]" strokeWidth={2.2} />}
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
								<FavoriteDmItem
									menuId="fav-self-dm"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									avatarSrc="/avatars/demo/marvell.png"
									label="ⱮąɾѵҽӀӀ"
									selected={isActive("fav-self-dm")}
									onSelect={selectNav("fav-self-dm", "ⱮąɾѵҽӀӀ", "dm")}
								/>
								<ChannelItem
									menuId="fav-announcements"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<SidebarListIconBox>{channelNavBellIcon}</SidebarListIconBox>}
									label="Announcements"
									selected={isActive("fav-announcements")}
									onSelect={selectNav("fav-announcements", "Announcements", "channel")}
								/>
								<ChannelItem
									menuId="fav-deployment-production"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<SidebarListIconBox>{channelNavBellIcon}</SidebarListIconBox>}
									label="Deployment Production"
									selected={isActive("fav-deployment-production")}
									onSelect={selectNav(
										"fav-deployment-production",
										"Deployment Production",
										"channel",
									)}
								/>
								<ChannelItem
									menuId="fav-deployment-staging"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<SidebarListIconBox>{channelNavBellIcon}</SidebarListIconBox>}
									label="Deployment Staging"
									selected={isActive("fav-deployment-staging")}
									onSelect={selectNav("fav-deployment-staging", "Deployment Staging", "channel")}
								/>
								<ChannelItem
									menuId="fav-release-notes"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<SidebarListIconBox>{channelNavHashIcon}</SidebarListIconBox>}
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
								<ChannelItem
									menuId="ch-infra"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<SidebarListIconBox>{channelNavInfraIcon}</SidebarListIconBox>}
									label="Infra"
									selected={isActive("ch-infra")}
									onSelect={selectNav("ch-infra", "Infra", "channel")}
								/>
								<ChannelItem
									menuId="ch-wichat"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<SidebarListIconBox>{channelNavWichatIcon}</SidebarListIconBox>}
									label="WiChat"
									selected={isActive("ch-wichat")}
									onSelect={selectNav("ch-wichat", "WiChat", "channel")}
								/>
								<ChannelItem
									menuId="ch-wicloud"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<SidebarListIconBox>{channelNavWicloudIcon}</SidebarListIconBox>}
									label="WiCloud"
									selected={isActive("ch-wicloud")}
									onSelect={selectNav("ch-wicloud", "WiCloud", "channel")}
								/>
								<ChannelItem
									menuId="ch-witask"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<SidebarListIconBox>{channelNavWitaskIcon}</SidebarListIconBox>}
									label="WiTask"
									selected={isActive("ch-witask")}
									onSelect={selectNav("ch-witask", "WiTask", "channel")}
								/>
								<ChannelItem
									menuId="ch-winobar"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<SidebarListIconBox>{channelNavWinobarIcon}</SidebarListIconBox>}
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
								<ChannelItem
									menuId="chat-dm-koh-yadi"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
									icon={<SidebarListAvatar src="/avatars/demo/koh-yadi.jpg" alt="Koh Yadi" />}
									label="Koh Yadi"
									selected={isActive("chat-dm-koh-yadi")}
									onSelect={selectNav("chat-dm-koh-yadi", "Koh Yadi", "dm")}
								/>
								<ChannelItem
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
					className="absolute top-0 right-0 z-20 h-full w-1.5 translate-x-1/2 cursor-col-resize touch-none hover:bg-primary/15 active:bg-primary/25"
					onPointerDown={onResizePointerDown}
				/>
			) : null}
		</div>
	);
}
