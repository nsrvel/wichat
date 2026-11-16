import { memo, type ReactNode } from "react";
import type { PresenceStatus } from "@wichat/contracts/user/presence";
import { OwnerCrownIcon } from "@/components/ui/owner-crown-icon";
import { UserAvatar } from "@/components/ui/user-avatar";
import { userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";

type MessageRowProps = {
	authorAvatarUrl?: string;
	authorDisplayName: string;
	authorNameColor?: string;
	authorIsOwner?: boolean;
	authorPresenceStatus?: PresenceStatus;
	ownerLabel?: string;
	timeLabel?: string;
	mine?: boolean;
	onAuthorClick?: () => void;
	children: ReactNode;
	reactions?: ReactNode;
};

export const MessageRow = memo(function MessageRow({
	authorAvatarUrl,
	authorDisplayName,
	authorNameColor,
	authorIsOwner = false,
	authorPresenceStatus,
	ownerLabel = "Owner",
	timeLabel,
	mine = false,
	onAuthorClick,
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
		<article className="flex gap-2.5 [content-visibility:auto] [contain-intrinsic-size:auto_68px]">
			{onAuthorClick ? (
				<button
					type="button"
					onClick={onAuthorClick}
					aria-label={authorDisplayName}
					className="mt-0.5 h-fit shrink-0 cursor-pointer rounded-full transition-opacity hover:opacity-85"
				>
					{avatarNode}
				</button>
			) : (
				avatarNode
			)}
			<div className="min-w-0 max-w-[min(88%,34rem)]">
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
				</div>
				<div className={bubbleClass}>{children}</div>
				{reactions ? (
					<div className="mt-1.5 flex flex-wrap items-center gap-1.5 px-0.5">{reactions}</div>
				) : null}
			</div>
		</article>
	);
});
