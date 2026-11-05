"use client";

import type { ComponentProps } from "react";
import { cn } from "@/lib/utils";

type WorkspaceSidebarResizeHandleProps = ComponentProps<"div"> & {
	sidebarWidth: number;
};

/** Full-height column resize — sits on the sidebar / chat boundary. */
export function WorkspaceSidebarResizeHandle({
	sidebarWidth,
	className,
	onPointerDown,
	...props
}: WorkspaceSidebarResizeHandleProps) {
	return (
		<div
			role="separator"
			aria-orientation="vertical"
			aria-label="Resize sidebar"
			className={cn(
				"absolute top-0 bottom-0 z-20 w-1.5 -translate-x-1/2 cursor-col-resize touch-none hover:bg-primary/15 active:bg-primary/25",
				className,
			)}
			style={{ left: sidebarWidth }}
			onPointerDown={onPointerDown}
			{...props}
		/>
	);
}
