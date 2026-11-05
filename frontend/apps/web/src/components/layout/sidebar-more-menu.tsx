"use client";

import { MoreHorizontal } from "lucide-react";
import { useEffect, useRef } from "react";
import { SidebarGhostIconButton } from "@/components/layout/sidebar-ghost-icon-button";
import { cn } from "@/lib/utils";

type SidebarMoreMenuProps = {
	id: string;
	activeId: string | null;
	onActiveIdChange: (id: string | null) => void;
	/** Tailwind classes controlling visibility (e.g. group-hover). */
	showClassName?: string;
};

export function SidebarMoreMenu({
	id,
	activeId,
	onActiveIdChange,
	showClassName,
}: SidebarMoreMenuProps) {
	const open = activeId === id;
	const ref = useRef<HTMLDivElement>(null);

	useEffect(() => {
		if (!open) return;
		const onDoc = (event: MouseEvent) => {
			if (ref.current?.contains(event.target as Node)) return;
			onActiveIdChange(null);
		};
		document.addEventListener("mousedown", onDoc);
		return () => document.removeEventListener("mousedown", onDoc);
	}, [open, onActiveIdChange]);

	return (
		<div
			ref={ref}
			className={cn("relative shrink-0 transition-opacity", showClassName, open && "opacity-100")}
		>
			<SidebarGhostIconButton
				type="button"
				aria-label="More actions"
				aria-expanded={open}
				className={open ? "bg-muted text-foreground" : undefined}
				onClick={(event) => {
					event.stopPropagation();
					onActiveIdChange(open ? null : id);
				}}
			>
				<MoreHorizontal className="scale-x-[1.2]" />
			</SidebarGhostIconButton>
			{open ? (
				<div
					className="absolute top-full right-0 z-50 mt-1 min-w-[9rem] rounded-lg border border-border bg-popover p-2 shadow-md"
					role="menu"
				>
					{/* Actions TBD */}
				</div>
			) : null}
		</div>
	);
}
