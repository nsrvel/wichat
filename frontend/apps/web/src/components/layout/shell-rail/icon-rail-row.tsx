import type { ComponentProps, ReactNode } from "react";
import { cn } from "@/lib/utils";

/** Horizontal inset — same on both sides inside the 64px rail. */
export const iconRailRowInsetClass = "mx-2";

type IconRailSelectionBarProps = {
	selected?: boolean;
	className?: string;
};

export function IconRailSelectionBar({ selected = false, className }: IconRailSelectionBarProps) {
	return (
		<span
			aria-hidden
			className={cn(
				"block w-1 shrink-0 rounded-r-full bg-sidebar-foreground transition-[height,opacity] duration-200",
				selected ? "h-6 opacity-100" : "h-2 opacity-40 group-hover:opacity-70",
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
				"group relative flex min-h-10 items-center justify-center",
				iconRailRowInsetClass,
				className,
			)}
		>
			<IconRailSelectionBar
				selected={selected}
				className="absolute top-1/2 -left-2 -translate-y-1/2"
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
			className={cn("flex min-h-10 items-center justify-center", iconRailRowInsetClass, className)}
			{...props}
		>
			{children}
		</div>
	);
}
