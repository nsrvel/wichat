"use client";

import type { ReactNode } from "react";
import { SidebarMoreMenu } from "@/components/layout/sidebar-more-menu";
import {
	sidebarRowMoreRailClass,
	sidebarRowShellClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { cn } from "@/lib/utils";

export type SidebarRowShellProps = {
	menuId: string;
	activeMenuId: string | null;
	onActiveMenuIdChange: (id: string | null) => void;
	unread?: boolean;
	selected?: boolean;
	rowClassName?: string;
	children: ReactNode;
};

export function SidebarRowShell({
	menuId,
	activeMenuId,
	onActiveMenuIdChange,
	unread,
	selected,
	rowClassName,
	children,
}: SidebarRowShellProps) {
	const menuOpen = activeMenuId === menuId;

	return (
		<div
			className={cn(sidebarRowShellClass, selected && "bg-sidebar-row-selected", rowClassName)}
			data-selected={selected ? "true" : undefined}
		>
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
