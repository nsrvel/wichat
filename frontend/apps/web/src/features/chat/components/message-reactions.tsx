"use client";

import { SmilePlus } from "lucide-react";
import { useTranslation } from "react-i18next";
import { ExpressionPickerPopover } from "@/components/expression-picker/expression-picker-popover";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { UserAvatar } from "@/components/ui/user-avatar";
import type { ChannelMessageReaction } from "@/features/chat/lib/channel-message";
import { userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";

export type ReactionUserInfo = {
	displayName: string;
	avatarUrl?: string;
};

type MessageReactionsProps = {
	reactions: ChannelMessageReaction[];
	className?: string;
	showAdd?: boolean;
	onToggleReaction?: (emoji: string) => void;
	resolveReactor?: (userId: string) => ReactionUserInfo | undefined;
};

export function MessageReactions({
	reactions,
	className,
	showAdd = true,
	onToggleReaction,
	resolveReactor,
}: MessageReactionsProps) {
	const { t } = useTranslation();

	if (reactions.length === 0) {
		return null;
	}

	return (
		<div className={cn("mt-1.5 flex flex-wrap items-center gap-1.5", className)}>
			{reactions.map((reaction) => {
				const userIds = reaction.userIds ?? [];
				const pillClass = cn(
					"flex h-6 min-w-[34px] items-center justify-center gap-1 rounded-full border px-2 text-xs transition active:scale-95",
					reaction.active
						? "border-primary/50 bg-primary/15 text-foreground"
						: "border-border/70 bg-muted/50 text-muted-foreground hover:border-border hover:bg-muted",
				);

				return (
					<Popover key={reaction.emoji}>
						<PopoverTrigger
							className={pillClass}
							aria-label={`${reaction.emoji} ${reaction.count ?? ""}`}
						>
							<span className="text-[13px] leading-none">{reaction.emoji}</span>
							{reaction.count !== undefined ? (
								<span className="text-[11px] font-semibold">{reaction.count}</span>
							) : null}
						</PopoverTrigger>
						<PopoverContent align="start" sideOffset={6} className="w-64 p-0">
							<div className="flex items-center justify-between border-b border-border px-3 py-2">
								<div className="flex items-center gap-1.5 text-xs font-semibold text-foreground">
									<span className="text-sm leading-none">{reaction.emoji}</span>
									<span>{reaction.count ?? userIds.length}</span>
								</div>
								{showAdd && onToggleReaction ? (
									<button
										type="button"
										onClick={() => onToggleReaction(reaction.emoji)}
										className={cn(
											"rounded-md px-2 py-0.5 text-[11px] font-medium transition",
											reaction.active
												? "bg-destructive/10 text-destructive hover:bg-destructive/20"
												: "bg-primary/15 text-primary hover:bg-primary/25",
										)}
									>
										{reaction.active
											? t("chat.reaction_remove_yours")
											: t("chat.reaction_add_yours")}
									</button>
								) : null}
							</div>
							<div className="thin-scroll max-h-56 overflow-y-auto p-1.5">
								{userIds.map((uid) => {
									const info = resolveReactor?.(uid);
									const name = info?.displayName || t("chat.reaction_unknown_user");
									return (
										<div
											key={uid}
											className="flex items-center justify-between gap-2 rounded-md px-2 py-1.5 text-xs hover:bg-accent/50"
										>
											<div className="flex min-w-0 items-center gap-2">
												<UserAvatar
													src={info?.avatarUrl}
													alt={name}
													initials={userInitials(name)}
													size="xs"
												/>
												<span className="truncate font-medium text-foreground">{name}</span>
											</div>
											<span className="shrink-0 text-sm leading-none">{reaction.emoji}</span>
										</div>
									);
								})}
							</div>
						</PopoverContent>
					</Popover>
				);
			})}
			{showAdd && onToggleReaction ? (
				<ExpressionPickerPopover
					only="emoji"
					closeOnSelect
					side="top"
					align="start"
					onSelect={(selection) => {
						if (selection.kind === "emoji") {
							onToggleReaction(selection.emoji);
						}
					}}
					trigger={
						<button
							type="button"
							className="flex h-6 w-7 items-center justify-center rounded-full border border-border/70 bg-muted/40 text-muted-foreground transition hover:bg-muted hover:text-foreground"
							aria-label={t("chat.add_reaction")}
							title={t("chat.add_reaction")}
						>
							<SmilePlus className="size-3.5" strokeWidth={1.8} />
						</button>
					}
				/>
			) : null}
		</div>
	);
}
