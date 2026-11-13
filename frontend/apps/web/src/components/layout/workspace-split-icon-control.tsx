import { ChevronDown } from "lucide-react";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

const segmentGhostClass =
	"bg-transparent text-muted-foreground hover:bg-sidebar-control hover:text-sidebar-foreground";
const segmentActiveClass =
	"bg-destructive/10 text-destructive hover:bg-destructive/20 hover:text-destructive dark:bg-destructive/20 dark:hover:bg-destructive/30";

type WorkspaceSplitIconControlProps = {
	mainAriaLabel: string;
	optionsAriaLabel: string;
	/** Muted / deafened — destructive fill; idle is ghost. */
	mainActive?: boolean;
	children: ReactNode;
};

/** Two adjacent icon buttons (outer corners rounded only), e.g. call / mic / headphones. */
export function WorkspaceSplitIconControl({
	mainAriaLabel,
	optionsAriaLabel,
	mainActive = false,
	children,
}: WorkspaceSplitIconControlProps) {
	const segmentClass = mainActive ? segmentActiveClass : segmentGhostClass;

	return (
		<div className="flex h-9 shrink-0 items-stretch">
			<button
				type="button"
				className={cn(
					"flex size-9 items-center justify-center rounded-l-lg transition-colors",
					segmentClass,
				)}
				aria-label={mainAriaLabel}
				aria-pressed={mainActive ? true : undefined}
			>
				{children}
			</button>
			<button
				type="button"
				className={cn(
					"flex h-9 shrink-0 items-center justify-center rounded-r-lg px-0.5 transition-colors",
					segmentClass,
				)}
				aria-label={optionsAriaLabel}
			>
				<ChevronDown className="size-3.5" strokeWidth={2.2} />
			</button>
		</div>
	);
}
