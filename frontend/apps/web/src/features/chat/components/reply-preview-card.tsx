"use client";

import { X } from "lucide-react";
import { useTranslation } from "react-i18next";
import { cn } from "@/lib/utils";

type ReplyPreviewCardProps = {
	messageId?: string;
	authorName: string;
	timeLabel?: string;
	bodyPreview: string;
	onClick?: () => void;
	onRemove?: () => void;
	className?: string;
};

export function ReplyPreviewCard({
	messageId,
	authorName,
	timeLabel,
	bodyPreview,
	onClick,
	onRemove,
	className,
}: ReplyPreviewCardProps) {
	const { t } = useTranslation();
	const interactive = Boolean(onClick);

	return (
		<div
			data-reply-jump-id={messageId || undefined}
			role={interactive ? "button" : undefined}
			tabIndex={interactive ? 0 : undefined}
			onClick={onClick}
			onKeyDown={
				interactive
					? (event) => {
							if (event.key === "Enter" || event.key === " ") {
								event.preventDefault();
								onClick?.();
							}
						}
					: undefined
			}
			className={cn(
				"flex w-fit max-w-lg items-start gap-2.5 rounded-lg border border-chat-bubble-border bg-chat-surface/65 px-2.5 py-2 text-left text-xs",
				interactive &&
					"cursor-pointer transition-colors hover:border-primary/40 hover:bg-chat-surface/85",
				className,
			)}
		>
			<span
				className="mt-0.5 w-[3px] shrink-0 self-stretch rounded-full bg-muted-foreground/55"
				aria-hidden
			/>
			<div className="min-w-0 flex-1">
				<div className="flex items-center gap-2 text-[11px] text-muted-foreground">
					<span className="truncate font-semibold">{authorName}</span>
					{timeLabel ? <span className="shrink-0">{timeLabel}</span> : null}
				</div>
				<p className="mt-0.5 line-clamp-2 break-words text-[13px] leading-snug text-foreground/95">
					{bodyPreview}
				</p>
			</div>
			{onRemove ? (
				<button
					type="button"
					onMouseDown={(event) => {
						event.preventDefault();
						event.stopPropagation();
					}}
					onClick={(event) => {
						event.stopPropagation();
						onRemove();
					}}
					className="ml-1 shrink-0 rounded p-0.5 text-muted-foreground transition hover:bg-sidebar-accent hover:text-foreground"
					aria-label={t("chat.cancel_reply")}
				>
					<X className="size-3.5" />
				</button>
			) : null}
		</div>
	);
}
