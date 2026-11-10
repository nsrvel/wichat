"use client";

import type { ReactNode } from "react";
import { SidebarRowShell } from "@/components/layout/channel-sidebar/sidebar-row-shell";
import {
	sidebarRowInnerButtonClass,
	sidebarRowTextClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { ChannelGlyphSlot } from "@/components/layout/sidebar-list-icon";
import { cn } from "@/lib/utils";

export type SidebarNavItemProps = {
	icon: ReactNode;
	label: string;
	badge?: boolean;
	unread?: boolean;
	selected?: boolean;
	onSelect?: () => void;
	menuId: string;
	activeMenuId: string | null;
	onActiveMenuIdChange: (id: string | null) => void;
};

export function SidebarNavItem({
	icon,
	label,
	badge,
	unread,
	selected,
	onSelect,
	menuId,
	activeMenuId,
	onActiveMenuIdChange,
}: SidebarNavItemProps) {
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
					selected ? "font-semibold text-sidebar-foreground" : undefined,
				)}
				onClick={onSelect}
			>
				<ChannelGlyphSlot framed={false}>{icon}</ChannelGlyphSlot>
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
