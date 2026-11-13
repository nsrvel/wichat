import Image from "next/image";
import { memo, type ReactNode } from "react";
import { userInitials } from "@/features/invite/lib/user-initials";
import { cn } from "@/lib/utils";

type MessageRowProps = {
	authorAvatarUrl?: string;
	authorDisplayName: string;
	timeLabel?: string;
	mine?: boolean;
	children: ReactNode;
	reactions?: ReactNode;
};

export const MessageRow = memo(function MessageRow({
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
			<article className="flex justify-end [content-visibility:auto] [contain-intrinsic-size:auto_64px]">
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
			</article>
		);
	}

	const hasAvatar = Boolean(authorAvatarUrl && authorAvatarUrl.trim().length > 0);

	return (
		<article className="flex gap-2.5 [content-visibility:auto] [contain-intrinsic-size:auto_68px]">
			{hasAvatar ? (
				<Image
					src={authorAvatarUrl!}
					alt={authorDisplayName}
					width={32}
					height={32}
					className="mt-0.5 size-8 shrink-0 rounded-full object-cover ring-1 ring-chat-surface-border"
				/>
			) : (
				<div
					className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full bg-background text-xs font-semibold text-foreground ring-1 ring-chat-surface-border"
					aria-hidden
				>
					{userInitials(authorDisplayName)}
				</div>
			)}
			<div className="min-w-0 max-w-[min(88%,34rem)]">
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
		</article>
	);
});
