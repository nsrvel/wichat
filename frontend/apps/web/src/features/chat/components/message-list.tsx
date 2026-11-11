"use client";

import type { ReactNode, RefObject } from "react";
import { cn } from "@/lib/utils";

export const MESSAGE_LIST_GUTTER_CLASS =
	"mx-auto flex w-full max-w-[56rem] flex-col gap-5 px-4 py-4 sm:px-5 lg:max-w-[60rem] lg:px-6";

type MessageListProps = {
	children: ReactNode;
	scrollRef?: RefObject<HTMLDivElement | null>;
	className?: string;
	innerClassName?: string;
};

export function MessageList({ children, scrollRef, className, innerClassName }: MessageListProps) {
	return (
		<div ref={scrollRef} className={cn("overlay-scroll flex-1 overflow-y-auto", className)}>
			<div className={cn(MESSAGE_LIST_GUTTER_CLASS, innerClassName)}>{children}</div>
		</div>
	);
}
