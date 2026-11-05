import type { ComponentProps, ReactNode } from "react";
import { cn } from "@/lib/utils";

type ConversationTabSelectionBarProps = {
	selected?: boolean;
	className?: string;
};

/** Thin underline on the header divider — color matches rail; shape stays flat, not the rail pill. */
export function ConversationTabSelectionBar({
	selected = false,
	className,
}: ConversationTabSelectionBarProps) {
	return (
		<span
			aria-hidden
			className={cn(
				"pointer-events-none block h-0.5 rounded-full bg-sidebar-foreground transition-opacity duration-200",
				selected ? "opacity-100" : "opacity-0",
				className,
			)}
		/>
	);
}

type ConversationTabRowProps = {
	selected?: boolean;
	children: ReactNode;
	className?: string;
};

/** Tab label centered in the header strip; indicator sits on the bottom edge. */
export function ConversationTabRow({
	selected = false,
	children,
	className,
}: ConversationTabRowProps) {
	return (
		<div className={cn("group relative flex h-full items-center justify-center px-3", className)}>
			<ConversationTabSelectionBar selected={selected} className="absolute inset-x-3 bottom-0" />
			{children}
		</div>
	);
}

type ConversationTabNavProps = ComponentProps<"div"> & {
	children: ReactNode;
};

/** Fills header strip height so tab rows can anchor the selection bar to the bottom divider. */
export function ConversationTabNav({ children, className, ...props }: ConversationTabNavProps) {
	return (
		<div className={cn("hidden min-w-0 flex-1 self-stretch md:flex", className)} {...props}>
			<nav className="flex h-full min-w-0 items-stretch" aria-label="Conversation sections">
				{children}
			</nav>
		</div>
	);
}

type ConversationTabActionRowProps = ComponentProps<"div"> & {
	children: ReactNode;
};

export function ConversationTabActionRow({
	children,
	className,
	...props
}: ConversationTabActionRowProps) {
	return (
		<div className={cn("flex self-center items-center pl-0.5", className)} {...props}>
			{children}
		</div>
	);
}
