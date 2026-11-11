import { cn } from "@/lib/utils";

export type ShellRailChromeState = "open" | "closed";

export function shellRailChromeState(open: boolean): ShellRailChromeState {
	return open ? "open" : "closed";
}

const shellRailMotion =
	"duration-[var(--shell-rail-motion-duration)] ease-[var(--shell-rail-motion-ease)] motion-reduce:transition-none";

/** Top bar + rail column share one `bg-shell-rail` surface (avoids a hairline between flex siblings). */
export function shellChromeBodyClass(className?: string) {
	return cn("flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden bg-shell-rail", className);
}

export function shellRailTrackClass(className?: string) {
	return cn("flex min-h-0 min-w-0 flex-1 overflow-hidden", className);
}

export function shellRailColumnClass(open: boolean) {
	return cn(
		"h-full min-h-0 shrink-0 overflow-hidden",
		"transition-[width]",
		shellRailMotion,
		open ? "w-[var(--shell-rail-width)]" : "w-0",
	);
}

/** `group/rail` on the aside — inner rows use `shellRailAsideInnerClass`. */
export function shellRailAsideClass(className?: string) {
	return cn(
		"group/rail flex h-full min-h-0 w-[var(--shell-rail-width)] flex-col gap-3 overflow-hidden bg-shell-rail pt-3",
		"transition-[padding-bottom,visibility]",
		shellRailMotion,
		"data-[state=open]:visible",
		"data-[state=closed]:pointer-events-none data-[state=closed]:invisible data-[state=closed]:delay-[var(--shell-rail-motion-duration)]",
		className,
	);
}

export function shellRailAsideInnerClass(className?: string) {
	return cn(
		"flex flex-col items-center gap-2 px-[var(--shell-rail-row-inset)]",
		"motion-reduce:transition-none group-data-[state=closed]/rail:opacity-0",
		className,
	);
}

/** Main workspace surface: top rule always; left + rounded TL only when the icon rail is open. */
export function shellMainPaneChromeClass(railOpen: boolean, className?: string) {
	return cn(
		"border-t border-shell-rail-border bg-chat-surface",
		"transition-[border-top-left-radius,border-left-width]",
		shellRailMotion,
		railOpen ? "rounded-tl-xl border-l" : "rounded-tl-none border-l-0",
		className,
	);
}
