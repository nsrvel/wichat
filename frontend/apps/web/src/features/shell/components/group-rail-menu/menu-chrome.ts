import { cn } from "@/lib/utils";

/** Above sidebar resize handle (z-10); matches dialog tokens at slightly smaller scale. */
export const menuPositionerClass = "z-[100] outline-none";

export const menuPopupClass = cn(
	"z-[100] w-max min-w-[13rem] max-w-[14.5rem] rounded-lg border border-border bg-card px-2 py-2.5 text-[13px] font-medium leading-snug",
	"text-card-foreground shadow-elevated outline-none",
);

export const menuItemClass = cn(
	"flex w-full cursor-default select-none items-center gap-2 rounded-md px-2.5 py-2.5 text-left outline-none",
	"whitespace-normal break-words text-card-foreground",
	"data-highlighted:bg-sidebar-accent data-highlighted:text-sidebar-accent-foreground",
);

export const menuGroupLabelClass =
	"block px-3 py-2 text-left text-xs font-normal leading-snug whitespace-normal break-words text-muted-foreground";

export const menuSeparatorClass = "mx-2.5 my-2 h-px bg-border/70";

export const menuHintClass = "text-[11px] font-normal leading-snug text-muted-foreground";

export const submenuTriggerClass = cn(menuItemClass, "justify-between gap-2.5 py-2.5 pr-1.5");

export const menuChevronClass = "size-3.5 shrink-0 text-muted-foreground/80";

/** Chevron on compact dropdown triggers (group header, invite member role, …). */
export const menuDropdownTriggerChevronClass = "size-3.5 shrink-0 text-muted-foreground";

/** Matches `WorkspaceGroupHeaderMenu` title dropdown trigger. */
export const menuDropdownTriggerClass = cn(
	"flex h-fit w-fit max-w-full min-w-0 items-center gap-1.5 rounded-md px-2 py-1.5 text-left transition-colors hover:bg-sidebar-accent",
	"min-h-[34px] lg:min-h-[36px] lg:h-[36px]",
);

export const menuDropdownTriggerOpenClass = "bg-sidebar-accent";

export const menuDropdownTriggerLabelClass =
	"truncate text-sm font-semibold tracking-tight text-foreground";

export const menuCheckboxItemClass = cn(
	menuItemClass,
	"items-center justify-between gap-3 py-2.5 pl-2.5 pr-2",
);

export const menuCheckboxBoxClass = cn(
	"flex size-4 shrink-0 items-center justify-center rounded-[3px] border border-border bg-muted/30",
	"group-data-checked/checkbox:border-primary group-data-checked/checkbox:bg-primary group-data-checked/checkbox:text-primary-foreground",
);

export const menuRadioItemClass = cn(menuCheckboxItemClass, "group/radio");

export const menuRadioCircleClass = cn(
	"relative box-border size-4 shrink-0 rounded-full border border-border bg-muted/30",
	"group-data-checked/radio:border-primary",
	"after:pointer-events-none after:absolute after:left-1/2 after:top-1/2 after:size-2 after:-translate-x-1/2 after:-translate-y-1/2",
	"after:rounded-full after:bg-primary after:opacity-0 after:content-['']",
	"group-data-checked/radio:after:opacity-100",
);

export const menuDestructiveItemClass = cn(
	menuItemClass,
	"text-destructive",
	"data-highlighted:bg-destructive/10 data-highlighted:text-destructive",
);
