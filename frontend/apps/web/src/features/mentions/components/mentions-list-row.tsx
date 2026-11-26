"use client";

import { Hash } from "lucide-react";
import { UserAvatar } from "@/components/ui/user-avatar";
import { MentionText } from "@/features/chat/lib/mention-text";
import { cn } from "@/lib/utils";

type MentionsListRowProps = {
	channelName: string;
	channelType: string;
	authorDisplayName: string;
	bodyPreview: string;
	createdAtLabel: string;
	isUnread?: boolean;
	selected?: boolean;
	selfUserId?: string;
	onClick: () => void;
};

export function MentionsListRow({
	channelName,
	channelType,
	authorDisplayName,
	bodyPreview,
	createdAtLabel,
	isUnread,
	selected,
	selfUserId,
	onClick,
}: MentionsListRowProps) {
	const isDm = channelType === "dm";

	return (
		<button
			type="button"
			onClick={onClick}
			className={cn(
				"flex w-full flex-col gap-1 border-b border-border/60 px-3 py-2.5 text-left transition hover:bg-muted/50",
				selected && "bg-chat-mention-self-bg",
				isUnread && !selected && "bg-chat-mention-self-bg/40",
			)}
		>
			<div className="flex items-center gap-2">
				{isDm ? (
					<UserAvatar src="" alt={channelName} initials={channelName.slice(0, 2)} size="xs" />
				) : (
					<span className="flex size-6 items-center justify-center rounded-md bg-muted text-muted-foreground">
						<Hash className="size-3.5" />
					</span>
				)}
				<span className="min-w-0 flex-1 truncate text-sm font-medium text-foreground">{channelName}</span>
				<span className="shrink-0 text-[11px] text-muted-foreground">{createdAtLabel}</span>
			</div>
			<p className="line-clamp-2 text-xs text-muted-foreground">
				<span className="font-medium text-foreground">{authorDisplayName}</span>
				{" — "}
				<MentionText text={bodyPreview} selfUserId={selfUserId} />
			</p>
		</button>
	);
}