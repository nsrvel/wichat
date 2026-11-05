"use client";

import { ChevronDown } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useShellChrome } from "@/components/layout/shell-chrome-context";
import { SidebarGhostIconButton } from "@/components/layout/sidebar-ghost-icon-button";
import { cn } from "@/lib/utils";

/** Primary sidebar filters (always visible, one row). */
export type WorkspaceSidebarPrimaryFilterId = "unread" | "channels" | "chats";

export type WorkspaceSidebarFilterId = WorkspaceSidebarPrimaryFilterId | string;

/** `null` = no filter (all items). */
export type WorkspaceSidebarFilterValue = WorkspaceSidebarFilterId | null;

const PRIMARY_FILTERS: { id: WorkspaceSidebarPrimaryFilterId; label: string }[] = [
	{ id: "unread", label: "Unread" },
	{ id: "channels", label: "Channels" },
	{ id: "chats", label: "Chats" },
];

/** Extra filters — product-defined; shown in chevron menu only. */
const MORE_FILTERS: { id: string; label: string }[] = [];

type WorkspaceSidebarFilterProps = {
	value?: WorkspaceSidebarFilterValue;
	onChange?: (id: WorkspaceSidebarFilterValue) => void;
	className?: string;
};

function FilterPill({
	label,
	selected,
	onClick,
	className,
}: {
	label: string;
	selected: boolean;
	onClick: () => void;
	className?: string;
}) {
	return (
		<button
			type="button"
			onClick={onClick}
			className={cn(
				"h-[22px] shrink-0 rounded-full px-2 text-[11px] font-medium leading-none transition-colors",
				selected
					? "bg-primary text-primary-foreground"
					: "border border-border bg-transparent text-muted-foreground hover:bg-muted",
				className,
			)}
		>
			{label}
		</button>
	);
}

export function WorkspaceSidebarFilter({
	value,
	onChange,
	className,
}: WorkspaceSidebarFilterProps) {
	const [internal, setInternal] = useState<WorkspaceSidebarFilterValue>(null);
	const [moreOpen, setMoreOpen] = useState(false);
	const moreRef = useRef<HTMLDivElement>(null);
	const { setSidebarFilterExpanded } = useShellChrome();
	const active = value !== undefined ? value : internal;

	useEffect(() => {
		setSidebarFilterExpanded(moreOpen);
	}, [moreOpen, setSidebarFilterExpanded]);

	const select = (id: WorkspaceSidebarFilterId) => {
		const next: WorkspaceSidebarFilterValue = active === id ? null : id;
		setInternal(next);
		onChange?.(next);
		setMoreOpen(false);
	};

	useEffect(() => {
		if (!moreOpen) return;
		const onDoc = (event: MouseEvent) => {
			if (moreRef.current?.contains(event.target as Node)) return;
			setMoreOpen(false);
		};
		document.addEventListener("mousedown", onDoc);
		return () => document.removeEventListener("mousedown", onDoc);
	}, [moreOpen]);

	const activeInMore = active !== null && MORE_FILTERS.some((f) => f.id === active);

	return (
		<div
			className={cn(
				"flex shrink-0 items-center gap-2 py-1.5 max-lg:border-b max-lg:border-chat-surface-border",
				className,
			)}
		>
			<div className="flex min-w-0 flex-1 flex-nowrap items-center gap-1 overflow-hidden">
				{PRIMARY_FILTERS.map((filter) => (
					<FilterPill
						key={filter.id}
						label={filter.label}
						selected={active === filter.id}
						onClick={() => select(filter.id)}
					/>
				))}
			</div>
			<div className="relative shrink-0" ref={moreRef}>
				<SidebarGhostIconButton
					type="button"
					aria-label="More filters"
					aria-expanded={moreOpen}
					className={cn((moreOpen || activeInMore) && "bg-muted text-foreground")}
					onClick={() => setMoreOpen((open) => !open)}
				>
					<ChevronDown
						className={cn("h-3.5 w-3.5 transition-transform", moreOpen && "rotate-180")}
					/>
				</SidebarGhostIconButton>
				{moreOpen && MORE_FILTERS.length > 0 ? (
					<div
						className="absolute top-full right-0 z-50 mt-1 min-w-[9rem] rounded-lg border border-border bg-popover p-1 shadow-md"
						role="menu"
					>
						{MORE_FILTERS.map((filter) => (
							<button
								key={filter.id}
								type="button"
								role="menuitem"
								className={cn(
									"flex w-full rounded-md px-2 py-1.5 text-left text-[11px] font-medium",
									active === filter.id
										? "bg-primary text-primary-foreground"
										: "text-foreground hover:bg-muted",
								)}
								onClick={() => select(filter.id)}
							>
								{filter.label}
							</button>
						))}
					</div>
				) : null}
			</div>
		</div>
	);
}
