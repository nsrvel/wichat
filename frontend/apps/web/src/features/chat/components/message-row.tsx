"use client";

import { memo, type ReactNode } from "react";
import type { MessageReplyPreview } from "@wichat/contracts/chat/message";
import type { PresenceStatus } from "@wichat/contracts/user/presence";
import { Pin } from "lucide-react";
import { useTranslation } from "react-i18next";
import { OwnerCrownIcon } from "@/components/ui/owner-crown-icon";
import { UserAvatar } from "@/components/ui/user-avatar";
import { ChatComposer } from "@/features/chat/components/chat-composer";
import { ReplyPreviewCard } from "@/features/chat/components/reply-preview-card";
import { userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";

type MessageRowProps = {
	messageId?: string;
	authorAvatarUrl?: string;
	authorDisplayName: string;
	authorNameColor?: string;
	authorIsOwner?: boolean;
	authorPresenceStatus?: PresenceStatus;
	ownerLabel?: string;
	timeLabel?: string;
	mine?: boolean;
	edited?: boolean;
	isPinned?: boolean;
	highlighted?: boolean;
	compact?: boolean;
	replyTo?: MessageReplyPreview | null;
	hasInlineReply?: boolean;
	onReplyClick?: (messageId: string) => void;
	onAuthorClick?: () => void;
	isEditing?: boolean;
	initialEditText?: string;
	onSaveEdit?: (nextBody: string) => void;
	onCancelEdit?: () => void;
	actionBar?: ReactNode;
	renderBubbleWrapper?: (bubble: ReactNode) => ReactNode;
	children: ReactNode;
	reactions?: ReactNode;
};

export const MessageRow = memo(function MessageRow({
	messageId,
	authorAvatarUrl,
	authorDisplayName,
	authorNameColor,
	authorIsOwner = false,
	authorPresenceStatus,
	ownerLabel = "Owner",
	timeLabel,
	mine = false,
	edited = false,
	isPinned = false,
	highlighted = false,
	compact = false,
	replyTo,
	hasInlineReply = false,
	onReplyClick,
	onAuthorClick,
	isEditing = false,
	initialEditText = "",
	onSaveEdit,
	onCancelEdit,
	actionBar,
	renderBubbleWrapper,
	children,
	reactions,
}: MessageRowProps) {
	const { t } = useTranslation();

	const bubbleClass = cn(
		"rounded-2xl border px-3.5 py-2.5 text-sm leading-[1.45] text-foreground transition-[box-shadow,background-color] duration-300 [&_p+p]:mt-2",
		mine
			? "rounded-tr-md border-chat-bubble-outgoing-border bg-chat-bubble-outgoing"
			: "rounded-tl-md border-chat-bubble-border bg-chat-bubble",
		isPinned && "ring-1 ring-amber-500/40",
		highlighted && "ring-2 ring-primary/70 shadow-sm",
	);

	const bodyContent = (
		<>
			{replyTo && !hasInlineReply ? (
				<div className="mb-1.5">
					<ReplyPreviewCard
						messageId={replyTo.id}
						authorName={replyTo.author_display_name}
						bodyPreview={replyTo.body_preview}
						onClick={onReplyClick ? () => onReplyClick(replyTo.id) : undefined}
					/>
				</div>
			) : null}
			{children}
			{edited ? (
				<span className="mt-1 block text-right text-[10px] text-muted-foreground">
					{t("chat.edited_tag")}
				</span>
			) : null}
		</>
	);

	const bubbleElement = <div className={bubbleClass}>{bodyContent}</div>;
	const wrappedBubble = isEditing ? (
		<ChatComposer
			mode="edit"
			initialValue={initialEditText}
			onSubmit={(nextBody) => onSaveEdit?.(nextBody)}
			onCancel={onCancelEdit}
			className="w-[min(82vw,34rem)]"
		/>
	) : renderBubbleWrapper ? (
		renderBubbleWrapper(bubbleElement)
	) : (
		bubbleElement
	);

	if (mine) {
		return (
			<article
				data-message-id={messageId}
				className={cn("group relative flex justify-end", compact ? "mt-1.5" : "mt-4")}
			>
				<div className="min-w-0 max-w-[min(88%,50rem)]">
					{!compact && (timeLabel || isPinned) ? (
						<div className="mb-1 flex items-center justify-end gap-1.5 px-0.5">
							{isPinned ? (
								<span className="inline-flex items-center gap-0.5 rounded bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-medium text-amber-600 dark:text-amber-400">
									<Pin className="size-2.5" />
								</span>
							) : null}
							{timeLabel ? (
								<span className="text-[11px] text-muted-foreground">{timeLabel}</span>
							) : null}
						</div>
					) : null}
					<div className="relative ml-auto w-fit max-w-full">
						{actionBar && !isEditing ? (
							<div className="pointer-events-none absolute -top-10 right-0 z-10 pb-1.5 opacity-0 transition-opacity group-focus-within:pointer-events-auto group-focus-within:opacity-100 group-hover:pointer-events-auto group-hover:opacity-100">
								{actionBar}
							</div>
						) : null}
						{wrappedBubble}
					</div>
					{reactions ? (
						<div className="mt-1.5 flex flex-wrap justify-end gap-1.5 px-0.5">{reactions}</div>
					) : null}
				</div>
			</article>
		);
	}

	const avatarNode = (
		<UserAvatar
			src={authorAvatarUrl}
			alt={authorDisplayName}
			initials={userInitials(authorDisplayName)}
			size="sm"
			status={authorPresenceStatus}
			statusRingClassName="ring-chat-surface"
			className={onAuthorClick ? undefined : "mt-0.5"}
		/>
	);

	return (
		<article
			data-message-id={messageId}
			className={cn("group relative flex gap-2.5", compact ? "mt-1.5" : "mt-4")}
		>
			{compact ? (
				<div className="w-8 shrink-0" aria-hidden />
			) : onAuthorClick ? (
				<button
					type="button"
					onClick={onAuthorClick}
					aria-label={authorDisplayName}
					className="mt-0.5 inline-flex h-fit shrink-0 cursor-pointer rounded-full transition-opacity hover:opacity-85"
				>
					{avatarNode}
				</button>
			) : (
				avatarNode
			)}

			<div className="min-w-0 max-w-[min(88%,50rem)]">
				{!compact ? (
					<div className="mb-1 flex items-center gap-1.5 px-0.5">
						{onAuthorClick ? (
							<button
								type="button"
								onClick={onAuthorClick}
								className="cursor-pointer text-[13px] font-semibold text-foreground hover:underline"
								style={authorNameColor ? { color: authorNameColor } : undefined}
							>
								{authorDisplayName}
							</button>
						) : (
							<span
								className="text-[13px] font-semibold text-foreground"
								style={authorNameColor ? { color: authorNameColor } : undefined}
							>
								{authorDisplayName}
							</span>
						)}
						{authorIsOwner ? <OwnerCrownIcon label={ownerLabel} /> : null}
						{timeLabel ? (
							<span className="ml-0.5 text-[11px] text-muted-foreground">{timeLabel}</span>
						) : null}
						{isPinned ? (
							<span className="inline-flex items-center gap-0.5 rounded bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-medium text-amber-600 dark:text-amber-400">
								<Pin className="size-2.5" />
							</span>
						) : null}
					</div>
				) : null}
				<div className="relative w-fit max-w-full">
					{actionBar && !isEditing ? (
						<div className="pointer-events-none absolute -top-10 left-0 z-10 pb-1.5 opacity-0 transition-opacity group-focus-within:pointer-events-auto group-focus-within:opacity-100 group-hover:pointer-events-auto group-hover:opacity-100">
							{actionBar}
						</div>
					) : null}
					{wrappedBubble}
				</div>
				{reactions ? (
					<div className="mt-1.5 flex flex-wrap items-center gap-1.5 px-0.5">{reactions}</div>
				) : null}
			</div>
		</article>
	);
});
