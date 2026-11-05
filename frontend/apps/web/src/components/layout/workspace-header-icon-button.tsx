import type { ComponentProps, ReactNode } from "react";
import { cn } from "@/lib/utils";

type WorkspaceHeaderIconButtonProps = ComponentProps<"button"> & {
	children: ReactNode;
};

/** 36×36 icon control — workspace sidebar header & shell top bar. */
export function WorkspaceHeaderIconButton({
	className,
	children,
	...props
}: WorkspaceHeaderIconButtonProps) {
	return (
		<button
			type="button"
			className={cn(
				"flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground",
				className,
			)}
			{...props}
		>
			{children}
		</button>
	);
}

export const workspaceHeaderIconClass = "h-[17px] w-[17px]";
