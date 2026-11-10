"use client";

import { ChevronDown } from "lucide-react";
import {
	sidebarCategoryInnerButtonClass,
	sidebarRowMoreRailClass,
	sidebarRowShellClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { SidebarMoreMenu } from "@/components/layout/sidebar-more-menu";
import { sidebarListIconSlotSizeClass } from "@/components/layout/sidebar-list-icon";
import { cn } from "@/lib/utils";

export type SidebarCategoryRowProps = {
	title: string;
	open: boolean;
	onToggle: () => void;
	menuId: string;
	activeMenuId: string | null;
	onActiveMenuIdChange: (id: string | null) => void;
};

export function SidebarCategoryRow({
	title,
	open,
	onToggle,
	menuId,
	activeMenuId,
	onActiveMenuIdChange,
}: SidebarCategoryRowProps) {
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
						"flex items-center justify-center text-muted-foreground",
						sidebarListIconSlotSizeClass,
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
