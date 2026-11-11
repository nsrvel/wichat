import Image from "next/image";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

type MessageRowProps = {
	authorAvatarUrl: string;
	authorDisplayName: string;
	timeLabel?: string;
	mine?: boolean;
	children: ReactNode;
	reactions?: ReactNode;
};

export function MessageRow({
	authorAvatarUrl,
	authorDisplayName,
	timeLabel,
	mine = false,
	children,
	reactions,
}: MessageRowProps) {
	const bubbleClass = cn(
		"rounded-2xl border px-3.5 py-2.5 text-sm leading-[1.45] text-foreground [&_p+p]:mt-2",
		mine
			? "rounded-tr-md border-chat-bubble-outgoing-border bg-chat-bubble-outgoing"
			: "rounded-tl-md border-chat-bubble-border bg-chat-bubble",
	);

	if (mine) {
		return (
			<div className="flex justify-end">
				<div className="min-w-0 max-w-[min(88%,34rem)]">
					{timeLabel ? (
						<div className="mb-1 flex justify-end px-0.5">
							<span className="text-[11px] text-muted-foreground">{timeLabel}</span>
						</div>
					) : null}
					<div className={bubbleClass}>{children}</div>
					{reactions ? (
						<div className="mt-1.5 flex flex-wrap justify-end gap-1.5 px-0.5">{reactions}</div>
					) : null}
				</div>
			</div>
		);
	}

	return (
		<div className="flex gap-2.5">
			<Image
				src={authorAvatarUrl}
				alt={authorDisplayName}
				width={32}
				height={32}
				className="mt-0.5 h-8 w-8 shrink-0 rounded-full object-cover ring-1 ring-chat-surface-border"
			/>
			<div className="min-w-0 max-w-full flex-1">
				<div className="mb-1 flex items-baseline gap-2 px-0.5">
					<span className="text-[13px] font-semibold text-foreground">{authorDisplayName}</span>
					{timeLabel ? (
						<span className="text-[11px] text-muted-foreground">{timeLabel}</span>
					) : null}
				</div>
				<div className={bubbleClass}>{children}</div>
				{reactions ? (
					<div className="mt-1.5 flex flex-wrap items-center gap-1.5 px-0.5">{reactions}</div>
				) : null}
			</div>
		</div>
	);
}
