"use client";

import type { Virtualizer } from "@tanstack/react-virtual";
import type { ReactNode, RefObject } from "react";
import { cn } from "@/lib/utils";

export const MESSAGE_LIST_GUTTER_CLASS =
	"mx-auto flex w-full max-w-4xl flex-col px-4 py-4 sm:px-5 lg:max-w-5xl lg:px-6";

type MessageListProps = {
	children?: ReactNode;
	header?: ReactNode;
	scrollRef?: RefObject<HTMLDivElement | null>;
	contentRef?: RefObject<HTMLDivElement | null>;
	onScroll?: () => void;
	className?: string;
	innerClassName?: string;
	virtualizer?: Virtualizer<HTMLDivElement, HTMLDivElement>;
	renderVirtualItem?: (index: number) => ReactNode;
};

export function MessageList({
	children,
	header,
	scrollRef,
	contentRef,
	onScroll,
	className,
	innerClassName,
	virtualizer,
	renderVirtualItem,
}: MessageListProps) {
	if (virtualizer && renderVirtualItem) {
		const virtualItems = virtualizer.getVirtualItems();
		const totalSize = virtualizer.getTotalSize();

		return (
			<div
				ref={scrollRef}
				onScroll={onScroll}
				className={cn(
					"overlay-scroll min-h-0 flex-1 overflow-y-auto [scrollbar-gutter:stable]",
					className,
				)}
			>
				<div ref={contentRef} className={cn(MESSAGE_LIST_GUTTER_CLASS, innerClassName)}>
					{header}
					{virtualItems.length > 0 ? (
						<div
							style={{
								height: `${totalSize}px`,
								width: "100%",
								position: "relative",
							}}
						>
							{virtualItems.map((virtualRow) => (
								<div
									key={virtualRow.key}
									data-index={virtualRow.index}
									ref={virtualizer.measureElement}
									style={{
										position: "absolute",
										top: 0,
										left: 0,
										width: "100%",
										transform: `translateY(${virtualRow.start}px)`,
									}}
								>
									{renderVirtualItem(virtualRow.index)}
								</div>
							))}
						</div>
					) : (
						children
					)}
				</div>
			</div>
		);
	}

	return (
		<div
			ref={scrollRef}
			onScroll={onScroll}
			className={cn(
				"overlay-scroll min-h-0 flex-1 overflow-y-auto [scrollbar-gutter:stable]",
				className,
			)}
		>
			<div ref={contentRef} className={cn(MESSAGE_LIST_GUTTER_CLASS, innerClassName)}>
				{header}
				{children}
			</div>
		</div>
	);
}
