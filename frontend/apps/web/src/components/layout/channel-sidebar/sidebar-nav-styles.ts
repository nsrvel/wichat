import { cn } from "@/lib/utils";

/** Horizontal inset for header + nav. */
export const SIDEBAR_INSET_X = "px-3.5";

const SIDEBAR_ROW_LG_H = "lg:min-h-[36px] lg:h-[36px]";

export function sidebarRowLabelClass(unread?: boolean) {
	return unread ? "font-semibold text-sidebar-foreground" : "font-medium text-muted-foreground";
}

export const sidebarRowShellClass = cn(
	"group/row relative flex min-h-[34px] items-center rounded-md transition-colors hover:bg-sidebar-accent",
	SIDEBAR_ROW_LG_H,
);

export function sidebarRowMoreRailClass(menuOpen: boolean) {
	return cn(
		"absolute top-0 right-0 z-[1] flex h-full w-8 items-center justify-end pr-0.5 opacity-0 transition-opacity group-hover/row:opacity-100",
		menuOpen && "opacity-100",
	);
}

export const sidebarRowInnerButtonClass = cn(
	"flex h-11 w-full min-w-0 items-center gap-3 rounded-md px-2.5 lg:gap-2 lg:px-1.5",
	SIDEBAR_ROW_LG_H,
);

/** Category header: same row height, chevron flush left (items indent via nested list). */
export const sidebarCategoryInnerButtonClass = cn(
	"flex h-11 w-full min-w-0 flex-1 items-center gap-1 rounded-md pl-0.5 pr-2.5",
	SIDEBAR_ROW_LG_H,
);

export const sidebarRowTextClass = "truncate text-sm lg:text-sm";

export const sidebarNestedListClass = "flex flex-col gap-0 pl-3 lg:pl-2.5";
