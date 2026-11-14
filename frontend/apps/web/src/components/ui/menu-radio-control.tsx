import { menuRadioCircleClass } from "@/lib/menu-chrome";
import { cn } from "@/lib/utils";

/** Visual radio control — selection state comes from parent `RadioItem` (`group/radio` + `data-checked`). */
export function MenuRadioControl({ className }: { className?: string }) {
	return <span className={cn(menuRadioCircleClass, className)} aria-hidden />;
}
