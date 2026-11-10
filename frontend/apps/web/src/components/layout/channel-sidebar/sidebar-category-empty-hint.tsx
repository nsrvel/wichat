import type { ReactNode } from "react";
import { sidebarNestedListClass } from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { cn } from "@/lib/utils";

type SidebarCategoryEmptyHintProps = {
	children: ReactNode;
	className?: string;
};

export function SidebarCategoryEmptyHint({ children, className }: SidebarCategoryEmptyHintProps) {
	return (
		<p
			className={cn(
				sidebarNestedListClass,
				"py-1.5 text-[12px] leading-snug text-muted-foreground lg:text-[13px]",
				className,
			)}
		>
			{children}
		</p>
	);
}
