import type { ComponentProps, ReactNode } from "react";
import { cn } from "@/lib/utils";

/** Row chrome only — horizontal inset lives on `shellRailAsideInnerClass`. */
export const iconRailRowInsetClass = "";

type IconRailSelectionBarProps = {
	selected?: boolean;
	className?: string;
};

export function IconRailSelectionBar({ selected = false, className }: IconRailSelectionBarProps) {
	return (
		<span
			aria-hidden
			className={cn(
				"block w-1 shrink-0 rounded-r-full bg-sidebar-foreground transition-opacity duration-150",
				selected
					? "h-10 opacity-100"
					: "h-2.5 opacity-35 group-hover:h-6 group-hover:opacity-75",
				className,
			)}
		/>
	);
}

type IconRailWorkspaceRowProps = {
	selected?: boolean;
	children: ReactNode;
	className?: string;
};

/** Workspace tile centered in rail; selection bar sits on the outer left edge. */
export function IconRailWorkspaceRow({
	selected = false,
	children,
	className,
}: IconRailWorkspaceRowProps) {
	return (
		<div
			className={cn(
				"group relative flex min-h-[var(--shell-rail-tile-size)] items-center justify-center",
				iconRailRowInsetClass,
				className,
			)}
		>
			<IconRailSelectionBar
				selected={selected}
				className="absolute top-1/2 -left-[var(--shell-rail-row-inset)] -translate-y-1/2"
			/>
			{children}
		</div>
	);
}

type IconRailActionRowProps = ComponentProps<"div"> & {
	children: ReactNode;
};

/** Add / utility tile — centered, no selection bar. */
export function IconRailActionRow({ children, className, ...props }: IconRailActionRowProps) {
	return (
		<div
			className={cn(
				"flex min-h-[var(--shell-rail-tile-size)] items-center justify-center",
				iconRailRowInsetClass,
				className,
			)}
			{...props}
		>
			{children}
		</div>
	);
}
