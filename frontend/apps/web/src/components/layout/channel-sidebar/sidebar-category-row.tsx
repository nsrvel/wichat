"use client";

import { ChevronDown, Plus } from "lucide-react";
import {
	sidebarCategoryInnerButtonClass,
	sidebarRowPersistentActionClass,
	sidebarRowShellClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { Button } from "@/components/ui/button";
import { sidebarListIconSlotSizeClass } from "@/components/layout/sidebar-list-icon";
import { cn } from "@/lib/utils";

export type SidebarCategoryRowProps = {
	title: string;
	open: boolean;
	onToggle: () => void;
	onAddClick?: () => void;
	addLabel?: string;
	/** @deprecated Retained only for legacy demo sidebar compatibility; unused. */
	menuId?: string;
	/** @deprecated Retained only for legacy demo sidebar compatibility; unused. */
	activeMenuId?: string | null;
	/** @deprecated Retained only for legacy demo sidebar compatibility; unused. */
	onActiveMenuIdChange?: (id: string | null) => void;
};

export function SidebarCategoryRow({
	title,
	open,
	onToggle,
	onAddClick,
	addLabel = "Create Channel",
}: SidebarCategoryRowProps) {
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
						className={cn("size-3.5 -rotate-90 transition-transform", open && "rotate-0")}
						strokeWidth={2.2}
					/>
				</span>
				<span className="min-w-0 flex-1 truncate text-xs font-medium text-muted-foreground transition-colors group-hover/row:text-foreground lg:text-[13px]">
					{title}
				</span>
			</button>
			{onAddClick ? (
				<div className={sidebarRowPersistentActionClass}>
					<Button
						type="button"
						variant="ghost"
						size="icon-sm"
						aria-label={addLabel}
						className="size-6 shrink-0 rounded-md p-0 text-muted-foreground hover:text-foreground"
						onClick={(event) => {
							event.stopPropagation();
							onAddClick();
						}}
					>
						<Plus className="size-3.5" aria-hidden />
					</Button>
				</div>
			) : null}
		</div>
	);
}
