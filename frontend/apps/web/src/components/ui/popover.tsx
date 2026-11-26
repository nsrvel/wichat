"use client";

import { Popover as PopoverPrimitive } from "@base-ui/react/popover";
import type { ComponentProps, ReactElement, ReactNode } from "react";
import { floatingPositionerDefaults } from "@/lib/floating-position";
import { cn } from "@/lib/utils";

type PopoverProps = {
	open?: boolean;
	defaultOpen?: boolean;
	onOpenChange?: ComponentProps<typeof PopoverPrimitive.Root>["onOpenChange"];
	modal?: ComponentProps<typeof PopoverPrimitive.Root>["modal"];
	children: ReactNode;
};

export function Popover({ open, defaultOpen, onOpenChange, modal, children }: PopoverProps) {
	return (
		<PopoverPrimitive.Root
			open={open}
			defaultOpen={defaultOpen}
			onOpenChange={onOpenChange}
			modal={modal}
		>
			{children}
		</PopoverPrimitive.Root>
	);
}

type PopoverTriggerProps = ComponentProps<typeof PopoverPrimitive.Trigger> & {
	render?: ReactElement;
	className?: string;
	children?: ReactNode;
};

export function PopoverTrigger({ render, className, children, ...rest }: PopoverTriggerProps) {
	return (
		<PopoverPrimitive.Trigger render={render} className={className} {...rest}>
			{children}
		</PopoverPrimitive.Trigger>
	);
}

type PopoverContentProps = {
	children: ReactNode;
	side?: ComponentProps<typeof PopoverPrimitive.Positioner>["side"];
	align?: ComponentProps<typeof PopoverPrimitive.Positioner>["align"];
	sideOffset?: number;
	anchor?: ComponentProps<typeof PopoverPrimitive.Positioner>["anchor"];
	className?: string;
};

export function PopoverContent({
	children,
	side = "top",
	align = "end",
	sideOffset = 10,
	anchor,
	className,
}: PopoverContentProps) {
	const floating = floatingPositionerDefaults();
	return (
		<PopoverPrimitive.Portal>
			<PopoverPrimitive.Positioner
				{...floating}
				anchor={anchor}
				side={side}
				align={align}
				sideOffset={sideOffset}
				className="z-[90]"
			>
				<PopoverPrimitive.Popup
					className={cn(
						"transform-gpu will-change-[transform,opacity] origin-[var(--transform-origin)] rounded-xl border border-border bg-popover text-popover-foreground shadow-xl outline-none",
						"transition-[transform,opacity] duration-100 ease-out data-[starting-style]:scale-[0.98] data-[starting-style]:opacity-0 data-[ending-style]:duration-75 data-[ending-style]:scale-[0.98] data-[ending-style]:opacity-0 motion-reduce:transition-none",
						className,
					)}
				>
					{children}
				</PopoverPrimitive.Popup>
			</PopoverPrimitive.Positioner>
		</PopoverPrimitive.Portal>
	);
}
