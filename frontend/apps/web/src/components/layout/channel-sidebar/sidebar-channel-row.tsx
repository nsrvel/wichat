"use client";

import type { ReactNode } from "react";
import { SidebarRowShell } from "@/components/layout/channel-sidebar/sidebar-row-shell";
import {
	sidebarRowInnerButtonClass,
	sidebarRowLabelClass,
	sidebarRowTextClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { cn } from "@/lib/utils";

export type SidebarChannelRowProps = {
	icon: ReactNode;
	label: string;
	unread?: boolean;
	selected?: boolean;
	menuId: string;
	activeMenuId: string | null;
	onActiveMenuIdChange: (id: string | null) => void;
	onSelect: () => void;
};

export function SidebarChannelRow({
	icon,
	label,
	unread,
	selected,
	menuId,
	activeMenuId,
	onActiveMenuIdChange,
	onSelect,
}: SidebarChannelRowProps) {
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
