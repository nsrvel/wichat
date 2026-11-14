"use client";

import { Popover as PopoverPrimitive } from "@base-ui/react/popover";
import type { ComponentProps, ReactElement, ReactNode } from "react";
import { cn } from "@/lib/utils";

type PopoverProps = {
	open?: boolean;
	defaultOpen?: boolean;
	onOpenChange?: (open: boolean) => void;
	children: ReactNode;
};

export function Popover({ open, defaultOpen, onOpenChange, children }: PopoverProps) {
	return (
		<PopoverPrimitive.Root
			open={open}
			defaultOpen={defaultOpen}
			onOpenChange={onOpenChange}
		>
			{children}
		</PopoverPrimitive.Root>
	);
}

type PopoverTriggerProps = {
	render?: ReactElement;
	className?: string;
	children?: ReactNode;
};

export function PopoverTrigger({ render, className, children }: PopoverTriggerProps) {
	return (
		<PopoverPrimitive.Trigger render={render} className={className}>
			{children}
		</PopoverPrimitive.Trigger>
	);
}

type PopoverContentProps = {
	children: ReactNode;
	side?: ComponentProps<typeof PopoverPrimitive.Positioner>["side"];
	align?: ComponentProps<typeof PopoverPrimitive.Positioner>["align"];
	sideOffset?: number;
	className?: string;
};

export function PopoverContent({
	children,
	side = "top",
	align = "end",
	sideOffset = 10,
	className,
}: PopoverContentProps) {
	return (
		<PopoverPrimitive.Portal>
			<PopoverPrimitive.Positioner
				side={side}
				align={align}
				sideOffset={sideOffset}
				className="z-[90]"
			>
				<PopoverPrimitive.Popup
					className={cn(
						"origin-[var(--transform-origin)] rounded-xl border border-border bg-popover text-popover-foreground shadow-xl outline-none",
						"transition-[transform,opacity] duration-150 ease-out data-[ending-style]:scale-95 data-[ending-style]:opacity-0 data-[starting-style]:scale-95 data-[starting-style]:opacity-0 motion-reduce:transition-none",
						className,
					)}
				>
					{children}
				</PopoverPrimitive.Popup>
			</PopoverPrimitive.Positioner>
		</PopoverPrimitive.Portal>
	);
}

