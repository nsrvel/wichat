"use client";

import type { ReactNode } from "react";
import {
	sidebarRowInnerButtonClass,
	sidebarRowMoreRailClass,
	sidebarRowShellClass,
	sidebarRowTextClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { ChannelGlyphSlot } from "@/components/layout/sidebar-list-icon";
import { cn } from "@/lib/utils";

export type SidebarPlainNavRowTrailingAction = {
	/** Accessible name for the trailing control (e.g. open in workspace). */
	label: string;
	icon: ReactNode;
	onClick: () => void;
	/** Default: show on row hover (channel-sidebar pattern). */
	visible?: "hover" | "always";
};

export type SidebarPlainNavRowProps = {
	icon: ReactNode;
	label: string;
	selected?: boolean;
	disabled?: boolean;
	onSelect?: () => void;
	trailingAction?: SidebarPlainNavRowTrailingAction;
};

/** Channel sidebar row without collapse chevron or row “more” menu. */
export function SidebarPlainNavRow({
	icon,
	label,
	selected,
	disabled,
	onSelect,
	trailingAction,
}: SidebarPlainNavRowProps) {
	const trailingAlways = trailingAction?.visible === "always";

	return (
		<div
			className={cn(sidebarRowShellClass, selected && "bg-sidebar-row-selected")}
			data-selected={selected ? "true" : undefined}
		>
			<button
				type="button"
				disabled={disabled}
				className={cn(
					sidebarRowInnerButtonClass,
					trailingAction && "pr-8",
					selected ? "font-semibold text-sidebar-foreground" : undefined,
					disabled && "cursor-not-allowed opacity-45",
				)}
				aria-current={selected ? "true" : undefined}
				onClick={onSelect}
			>
				<ChannelGlyphSlot framed={false}>{icon}</ChannelGlyphSlot>
				<span
					className={cn(
						sidebarRowTextClass,
						"min-w-0 flex-1 truncate text-left font-medium capitalize",
					)}
				>
					{label}
				</span>
			</button>
			{trailingAction ? (
				<div
					className={cn(sidebarRowMoreRailClass(false), trailingAlways ? "opacity-100" : undefined)}
				>
					<button
						type="button"
						disabled={disabled}
						className={cn(
							"flex size-7 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors",
							"hover:bg-sidebar-accent hover:text-sidebar-foreground",
							disabled && "pointer-events-none opacity-45",
						)}
						aria-label={trailingAction.label}
						onClick={(event) => {
							event.stopPropagation();
							trailingAction.onClick();
						}}
					>
						{trailingAction.icon}
					</button>
				</div>
			) : null}
		</div>
	);
}
