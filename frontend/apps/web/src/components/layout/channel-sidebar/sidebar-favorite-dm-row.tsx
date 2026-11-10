"use client";

import { SidebarRowShell } from "@/components/layout/channel-sidebar/sidebar-row-shell";
import {
	sidebarRowInnerButtonClass,
	sidebarRowLabelClass,
	sidebarRowTextClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { SidebarListAvatar } from "@/components/layout/sidebar-list-icon";
import { cn } from "@/lib/utils";

export type SidebarFavoriteDmRowProps = {
	avatarSrc: string;
	label: string;
	unread?: boolean;
	selected?: boolean;
	menuId: string;
	activeMenuId: string | null;
	onActiveMenuIdChange: (id: string | null) => void;
	onSelect: () => void;
};

export function SidebarFavoriteDmRow({
	avatarSrc,
	label,
	unread,
	selected,
	menuId,
	activeMenuId,
	onActiveMenuIdChange,
	onSelect,
}: SidebarFavoriteDmRowProps) {
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
