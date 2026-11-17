"use client";

import type { ReactNode, RefObject } from "react";
import { cn } from "@/lib/utils";

export const MESSAGE_LIST_GUTTER_CLASS =
	"mx-auto flex w-full max-w-4xl flex-col px-4 py-4 sm:px-5 lg:max-w-5xl lg:px-6";

type MessageListProps = {
	children: ReactNode;
	scrollRef?: RefObject<HTMLDivElement | null>;
	contentRef?: RefObject<HTMLDivElement | null>;
	onScroll?: () => void;
	className?: string;
	innerClassName?: string;
};

export function MessageList({
	children,
	scrollRef,
	contentRef,
	onScroll,
	className,
	innerClassName,
}: MessageListProps) {
	return (
		<div
			ref={scrollRef}
			onScroll={onScroll}
			style={{ overflowAnchor: "none" }}
			className={cn(
				"overlay-scroll min-h-0 flex-1 overflow-y-auto [scrollbar-gutter:stable]",
				className,
			)}
		>
			<div ref={contentRef} className={cn(MESSAGE_LIST_GUTTER_CLASS, innerClassName)}>
				{children}
			</div>
		</div>
	);
}
