"use client";

import type { ReactNode } from "react";
import { shellMainPaneChromeClass } from "@/components/layout/shell-rail/shell-rail-aside";
import { cn } from "@/lib/utils";

export type ShellMainPaneProps = {
	children?: ReactNode;
	railOpen?: boolean;
	className?: string;
};

/**
 * Shell **main pane** (workspace canvas): white chat surface on the gray shell strip.
 * Border top always; border left + `rounded-tl-xl` only when the icon rail is open. Route content or
 * `WorkspacePanel` (dev mock) renders inside — not a separate “canvas” component.
 */
export function ShellMainPane({ children, railOpen = true, className }: ShellMainPaneProps) {
	return (
		<div
			className={cn(
				"min-h-0 min-w-0 flex-1 overflow-hidden",
				shellMainPaneChromeClass(railOpen),
				className,
			)}
		>
			{children}
		</div>
	);
}
