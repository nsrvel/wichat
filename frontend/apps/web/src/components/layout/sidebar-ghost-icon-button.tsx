import type { ComponentProps } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const ghostIconClass =
	"size-7 shrink-0 rounded-md p-0 text-muted-foreground [&_svg]:size-[18px] [&_svg]:stroke-[2.1] [&_svg]:transition-[stroke-width] hover:[&_svg]:stroke-[2.75]";

type SidebarGhostIconButtonProps = ComponentProps<typeof Button>;

export function SidebarGhostIconButton({
	className,
	variant = "ghost",
	size = "icon-sm",
	...props
}: SidebarGhostIconButtonProps) {
	return (
		<Button variant={variant} size={size} className={cn(ghostIconClass, className)} {...props} />
	);
}
