"use client";

import {
	sidebarRowInnerButtonClass,
	sidebarRowShellClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { cn } from "@/lib/utils";

export type SidebarStaticCategoryRowProps = {
	title: string;
};

/** Category label matching channel sidebar chrome — always visible, no collapse. */
export function SidebarStaticCategoryRow({ title }: SidebarStaticCategoryRowProps) {
	return (
		<div className={cn(sidebarRowShellClass, "pointer-events-none hover:bg-transparent")}>
			<div className={cn(sidebarRowInnerButtonClass, "pointer-events-none min-w-0 flex-1")}>
				<span className="min-w-0 flex-1 truncate text-left text-xs font-medium capitalize text-muted-foreground lg:text-[13px]">
					{title}
				</span>
			</div>
		</div>
	);
}
