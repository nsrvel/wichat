import { useState, type ComponentProps, type ReactNode } from "react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

const tileBase =
	"flex h-[var(--shell-rail-tile-size)] w-[var(--shell-rail-tile-size)] shrink-0 items-center justify-center rounded-[var(--shell-rail-tile-radius)] outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-shell-rail";

const tileVariants = {
	primary: "bg-[#141414] text-white shadow-sm",
	tile: "border border-shell-rail-border bg-shell-rail-tile text-[#1c1c1e] shadow-[0_1px_2px_rgba(0,0,0,0.06)] dark:text-[#f4f4f5]",
	muted: "bg-shell-rail-action text-[#3f3f46] hover:bg-shell-rail-action-hover dark:text-[#e4e4e7]",
} as const;

type IconRailTileButtonProps = ComponentProps<"button"> & {
	variant?: keyof typeof tileVariants;
	children: ReactNode;
};

export function IconRailTileButton({
	variant = "muted",
	className,
	children,
	...props
}: IconRailTileButtonProps) {
	return (
		<button type="button" className={cn(tileBase, tileVariants[variant], className)} {...props}>
			{children}
		</button>
	);
}

type IconRailTooltipTileProps = Omit<IconRailTileButtonProps, "aria-label"> & {
	/** Accessible name and tooltip label (required for icon-only controls). */
	label: string;
	tooltipSide?: ComponentProps<typeof TooltipContent>["side"];
};

export function IconRailTooltipTile({
	label,
	tooltipSide = "right",
	children,
	onClick,
	onPointerDown,
	...buttonProps
}: IconRailTooltipTileProps) {
	const [open, setOpen] = useState(false);

	return (
		<Tooltip open={open} onOpenChange={setOpen}>
			<TooltipTrigger
				delay={300}
				closeDelay={0}
				render={
					<IconRailTileButton
						aria-label={label}
						{...buttonProps}
						onPointerDown={(event) => {
							setOpen(false);
							onPointerDown?.(event);
						}}
						onClick={(event) => {
							setOpen(false);
							onClick?.(event);
						}}
					>
						{children}
					</IconRailTileButton>
				}
			/>
			<TooltipContent side={tooltipSide}>{label}</TooltipContent>
		</Tooltip>
	);
}
