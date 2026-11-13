"use client";

import { Check } from "lucide-react";
import type { ReactNode } from "react";
import { useGroupMenuKit } from "@/features/shell/components/group-rail-menu/group-menu-kit";
import {
	menuCheckboxBoxClass,
	menuCheckboxItemClass,
} from "@/features/shell/components/group-rail-menu/menu-chrome";
import { cn } from "@/lib/utils";

type GroupRailMenuCheckboxItemProps = {
	label: string;
	checked: boolean;
	onCheckedChange: (checked: boolean) => void;
	leadingIcon?: ReactNode;
};

export function GroupRailMenuCheckboxItem({
	label,
	checked,
	onCheckedChange,
	leadingIcon,
}: GroupRailMenuCheckboxItemProps) {
	const { CheckboxItem, CheckboxItemIndicator } = useGroupMenuKit();

	return (
		<CheckboxItem
			className={cn(menuCheckboxItemClass, "group/checkbox")}
			checked={checked}
			onCheckedChange={onCheckedChange}
			closeOnClick={false}
		>
			{leadingIcon}
			<span className="min-w-0 flex-1 pr-1">{label}</span>
			<span className={menuCheckboxBoxClass} aria-hidden>
				<CheckboxItemIndicator
					keepMounted
					className="flex items-center justify-center opacity-0 data-checked:opacity-100"
				>
					<Check className="size-3" strokeWidth={2.5} />
				</CheckboxItemIndicator>
			</span>
		</CheckboxItem>
	);
}
