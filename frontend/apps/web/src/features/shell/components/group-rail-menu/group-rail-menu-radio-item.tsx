"use client";

import { useGroupMenuKit } from "@/features/shell/components/group-rail-menu/group-menu-kit";
import { MenuRadioControl } from "@/components/ui/menu-radio-control";
import { menuHintClass, menuRadioItemClass } from "@/lib/menu-chrome";
import { cn } from "@/lib/utils";

type GroupRailMenuRadioItemProps = {
	label: string;
	value: string;
	hint?: string;
};

export function GroupRailMenuRadioItem({ label, value, hint }: GroupRailMenuRadioItemProps) {
	const { RadioItem } = useGroupMenuKit();

	return (
		<RadioItem className={cn(menuRadioItemClass, "group/radio")} value={value} closeOnClick={false}>
			{hint ? (
				<div className="flex min-w-0 flex-1 flex-col items-start gap-0.5 pr-1">
					<span className="leading-tight">{label}</span>
					<span className={menuHintClass}>{hint}</span>
				</div>
			) : (
				<span className="min-w-0 flex-1 pr-1">{label}</span>
			)}
			<MenuRadioControl />
		</RadioItem>
	);
}
