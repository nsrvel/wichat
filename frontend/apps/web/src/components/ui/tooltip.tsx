"use client";

import { Tooltip as TooltipPrimitive } from "@base-ui/react/tooltip";
import type { ComponentProps, ReactNode } from "react";
import { cn } from "@/lib/utils";

export function TooltipProvider({
	children,
	delay = 400,
	closeDelay = 0,
}: TooltipPrimitive.Provider.Props) {
	return (
		<TooltipPrimitive.Provider delay={delay} closeDelay={closeDelay}>
			{children}
		</TooltipPrimitive.Provider>
	);
}

type TooltipProps = {
	children: ReactNode;
	open?: boolean;
	onOpenChange?: (open: boolean) => void;
};

export function Tooltip({ children, open, onOpenChange }: TooltipProps) {
	return (
		<TooltipPrimitive.Root open={open} onOpenChange={onOpenChange}>
			{children}
		</TooltipPrimitive.Root>
	);
}

type TooltipContentProps = {
	children: ReactNode;
	side?: ComponentProps<typeof TooltipPrimitive.Positioner>["side"];
	sideOffset?: number;
	className?: string;
};

export function TooltipContent({
	children,
	side = "right",
	sideOffset = 8,
	className,
}: TooltipContentProps) {
	return (
		<TooltipPrimitive.Portal>
			<TooltipPrimitive.Positioner side={side} sideOffset={sideOffset}>
				<TooltipPrimitive.Popup
					className={cn(
						"z-50 max-w-xs rounded-md border border-border bg-popover px-2.5 py-1.5 text-xs font-medium text-popover-foreground shadow-md",
						"origin-[var(--transform-origin)] motion-reduce:transition-none",
						className,
					)}
				>
					{children}
				</TooltipPrimitive.Popup>
			</TooltipPrimitive.Positioner>
		</TooltipPrimitive.Portal>
	);
}

type TooltipTriggerProps = TooltipPrimitive.Trigger.Props;

export function TooltipTrigger(props: TooltipTriggerProps) {
	return <TooltipPrimitive.Trigger {...props} />;
}
