"use client";

import { useGroupMenuKit } from "@/features/shell/components/group-rail-menu/group-menu-kit";
import { MenuRadioControl } from "@/features/shell/components/group-rail-menu/menu-radio-control";
import { menuRadioItemClass } from "@/features/shell/components/group-rail-menu/menu-chrome";
import { cn } from "@/lib/utils";

type GroupRailMenuRadioItemProps = {
	label: string;
	value: string;
};

export function GroupRailMenuRadioItem({ label, value }: GroupRailMenuRadioItemProps) {
	const { RadioItem } = useGroupMenuKit();

	return (
		<RadioItem
			className={cn(menuRadioItemClass, "group/radio")}
			value={value}
			closeOnClick={false}
		>
			<span className="min-w-0 flex-1 pr-1">{label}</span>
			<MenuRadioControl />
		</RadioItem>
	);
}
