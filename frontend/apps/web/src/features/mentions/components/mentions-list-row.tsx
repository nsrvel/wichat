"use client";

import { ChannelListIcon } from "@/components/layout/channel-list-icon";
import { UserAvatar } from "@/components/ui/user-avatar";
import { MentionText } from "@/features/chat/lib/mention-text";
import { userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";

type MentionsListRowProps = {
	groupId?: string;
	channelName: string;
	channelType: string;
	channelImageUrl?: string;
	authorDisplayName: string;
	authorAvatarUrl?: string;
	bodyPreview: string;
	createdAtLabel: string;
	isUnread?: boolean;
	selected?: boolean;
	selfUserId?: string;
	onClick: () => void;
};

export function MentionsListRow({
	groupId,
	channelName,
	channelType,
	channelImageUrl,
	authorDisplayName,
	authorAvatarUrl,
	bodyPreview,
	createdAtLabel,
	isUnread,
	selected,
	selfUserId,
	onClick,
}: MentionsListRowProps) {
	const isDm = channelType === "dm";
	const dmAvatarSrc = isDm ? (authorAvatarUrl ?? "") : "";

	return (
		<button
			type="button"
			onClick={onClick}
			className={cn(
				"flex w-full items-start gap-3 rounded-lg px-2.5 py-2.5 text-left transition hover:bg-muted/60",
				selected && "bg-muted",
				isUnread && !selected && "bg-chat-mention-self-bg/35 hover:bg-chat-mention-self-bg/50",
			)}
		>
			{/* Left: larger channel / DM avatar */}
			<div className="mt-0.5 shrink-0">
				{isDm ? (
					<UserAvatar
						src={dmAvatarSrc}
						alt={channelName}
						initials={userInitials(channelName)}
						size="md"
					/>
				) : (
					<div className="flex size-9 items-center justify-center rounded-full bg-muted/80 [&>*]:!size-9 [&>*]:!rounded-full">
						<ChannelListIcon
							type={channelType}
							channelName={channelName}
							imageUrl={channelImageUrl}
							variant="header"
						/>
					</div>
				)}
			</div>

			{/* Right: channel title + date on top, sender avatar + name: message below */}
			<div className="min-w-0 flex-1 space-y-1">
				<div className="flex items-center justify-between gap-2">
					<span
						className={cn(
							"min-w-0 truncate text-[13.5px] leading-tight text-foreground",
							isUnread ? "font-semibold" : "font-medium",
						)}
					>
						{channelName}
					</span>
					<span className="flex shrink-0 items-center gap-1.5 text-[11px] tabular-nums text-muted-foreground">
						<span>{createdAtLabel}</span>
						{isUnread ? <span className="size-1.5 rounded-full bg-primary" aria-hidden /> : null}
					</span>
				</div>

				<div className="flex items-start gap-1.5">
					<UserAvatar
						src={authorAvatarUrl ?? ""}
						alt={authorDisplayName}
						initials={userInitials(authorDisplayName)}
						size="xs"
						className="mt-0.5 size-[18px] shrink-0 text-[9px]"
					/>
					<p className="line-clamp-2 min-w-0 flex-1 text-xs leading-snug text-muted-foreground">
						<span className="font-medium text-foreground/90">{authorDisplayName}:</span>{" "}
						<MentionText text={bodyPreview} selfUserId={selfUserId} groupId={groupId} />
					</p>
				</div>
			</div>
		</button>
	);
}
