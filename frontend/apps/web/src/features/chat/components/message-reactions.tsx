import { SmilePlus } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ChannelMessageReaction } from "@/features/chat/lib/channel-message";

type MessageReactionsProps = {
	reactions: ChannelMessageReaction[];
	className?: string;
	showAdd?: boolean;
};

export function MessageReactions({ reactions, className, showAdd = true }: MessageReactionsProps) {
	if (reactions.length === 0 && !showAdd) {
		return null;
	}
	return (
		<div className={cn("mt-1.5 flex flex-wrap items-center gap-1.5 px-0.5", className)}>
			{reactions.map((reaction) => (
				<span
					key={`${reaction.emoji}-${reaction.count ?? 0}`}
					className={cn(
						"flex h-6 min-w-[30px] items-center justify-center gap-1 rounded-full border px-1.5 text-xs",
						reaction.active
							? "border-[#c9d6ec] bg-[#f0f4fa] dark:border-[#3d4f6f] dark:bg-[#1e2a3d]"
							: "border-chat-surface-border bg-chat-composer-field",
					)}
				>
					<span className="text-[13px] leading-none">{reaction.emoji}</span>
					{reaction.count !== undefined ? (
						<span className="text-[11px] font-medium text-muted-foreground">{reaction.count}</span>
					) : null}
				</span>
			))}
			{showAdd ? (
				<button
					type="button"
					className="flex h-6 w-8 items-center justify-center rounded-full border border-chat-surface-border bg-chat-surface text-muted-foreground transition-colors hover:bg-sidebar-accent"
					aria-label="Add reaction"
				>
					<SmilePlus className="size-3.5" strokeWidth={1.8} />
				</button>
			) : null}
		</div>
	);
}
