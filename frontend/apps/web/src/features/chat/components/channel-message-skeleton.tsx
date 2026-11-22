"use client";

import { cn } from "@/lib/utils";
import { MESSAGE_LIST_GUTTER_CLASS } from "@/features/chat/components/message-list";

function SkeletonRow({ compact }: { compact?: boolean }) {
	return (
		<div className={cn("flex gap-3 py-1", compact ? "py-0.5" : "py-2")}>
			<div className="size-9 shrink-0 animate-pulse rounded-full bg-muted/80" />
			<div className="min-w-0 flex-1 space-y-2 pt-0.5">
				<div className="flex items-center gap-2">
					<div className="h-3 w-24 animate-pulse rounded bg-muted/80" />
					<div className="h-2.5 w-10 animate-pulse rounded bg-muted/60" />
				</div>
				<div className="h-3 w-[88%] max-w-md animate-pulse rounded bg-muted/70" />
				{!compact ? <div className="h-3 w-[62%] max-w-sm animate-pulse rounded bg-muted/60" /> : null}
			</div>
		</div>
	);
}

export function ChannelMessageSkeleton({ rows = 8 }: { rows?: number }) {
	return (
		<div className={MESSAGE_LIST_GUTTER_CLASS} aria-busy="true" aria-live="polite">
			{Array.from({ length: rows }, (_, i) => (
				<SkeletonRow key={i} compact={i > 0 && i % 3 !== 0} />
			))}
		</div>
	);
}