"use client";

import type { ComponentProps, ReactElement } from "react";
import { useState } from "react";
import {
	ExpressionPicker,
	type ExpressionPickerProps,
} from "@/components/expression-picker/expression-picker";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import type { ExpressionSelection } from "@/lib/expressions/expression-types";

type ExpressionPickerPopoverProps = ExpressionPickerProps & {
	trigger: ReactElement;
	open?: boolean;
	onOpenChange?: (open: boolean) => void;
	closeOnSelect?: boolean;
	side?: ComponentProps<typeof PopoverContent>["side"];
	align?: ComponentProps<typeof PopoverContent>["align"];
	sideOffset?: number;
};

export function ExpressionPickerPopover({
	trigger,
	open: controlledOpen,
	onOpenChange,
	closeOnSelect = false,
	side = "top",
	align = "end",
	sideOffset = 2,
	onSelect,
	...pickerProps
}: ExpressionPickerPopoverProps) {
	const [internalOpen, setInternalOpen] = useState(false);
	const isControlled = controlledOpen !== undefined;
	const open = isControlled ? controlledOpen : internalOpen;

	const setOpen = (next: boolean) => {
		if (!isControlled) {
			setInternalOpen(next);
		}
		onOpenChange?.(next);
	};

	const handleSelect = (selection: ExpressionSelection) => {
		onSelect(selection);
		if (closeOnSelect || selection.kind !== "emoji") {
			setOpen(false);
		}
	};

	return (
		<Popover open={open} onOpenChange={setOpen}>
			<PopoverTrigger render={trigger} />
			<PopoverContent
				side={side}
				align={align}
				sideOffset={sideOffset}
				className="border border-border bg-card p-0 text-card-foreground shadow-[0_14px_40px_-8px_rgba(0,0,0,0.45)]"
			>
				{open ? <ExpressionPicker {...pickerProps} onSelect={handleSelect} /> : null}
			</PopoverContent>
		</Popover>
	);
}
