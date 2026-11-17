"use client";

import { Pencil, Reply, SmilePlus } from "lucide-react";
import { useTranslation } from "react-i18next";
import { ExpressionPickerPopover } from "@/components/expression-picker/expression-picker-popover";
import {
	MessageMoreMenuButton,
	type MessageMenuActionProps,
} from "@/features/chat/components/message-context-menu";
import { cn } from "@/lib/utils";

const QUICK_EMOJIS = ["👍", "❤️", "😂", "😮"] as const;

type MessageActionBarProps = MessageMenuActionProps & {
	canReact: boolean;
	onReact: (emoji: string) => void;
};

export function MessageActionBar({
	isMine,
	isPinned,
	canReact,
	canPin,
	canDelete,
	onReact,
	onReply,
	onEdit,
	onTogglePin,
	onCopy,
	onDelete,
}: MessageActionBarProps) {
	const { t } = useTranslation();

	const menuProps: MessageMenuActionProps = {
		isMine,
		isPinned,
		canPin,
		canDelete,
		onReply,
		onEdit,
		onTogglePin,
		onCopy,
		onDelete,
	};

	return (
		<div
			className={cn(
				"flex items-center gap-0.5 rounded-lg border border-border bg-card p-1 text-card-foreground shadow-elevated",
			)}
		>
			{canReact ? (
				<>
					{QUICK_EMOJIS.map((emoji) => (
						<button
							key={emoji}
							type="button"
							onClick={() => onReact(emoji)}
							className="flex size-7 items-center justify-center rounded-md text-sm transition hover:bg-sidebar-accent"
							aria-label={`${t("chat.add_reaction")} ${emoji}`}
						>
							{emoji}
						</button>
					))}
					<ExpressionPickerPopover
						only="emoji"
						closeOnSelect
						side="top"
						align="end"
						onSelect={(selection) => {
							if (selection.kind === "emoji") {
								onReact(selection.emoji);
							}
						}}
						trigger={
							<button
								type="button"
								className="flex size-7 items-center justify-center rounded-md text-muted-foreground transition hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
								aria-label={t("chat.add_reaction")}
								title={t("chat.add_reaction")}
							>
								<SmilePlus className="size-4" />
							</button>
						}
					/>
					<span className="mx-0.5 h-4 w-px bg-border/70" aria-hidden />
				</>
			) : null}

			{isMine && onEdit ? (
				<button
					type="button"
					onClick={onEdit}
					className="flex size-7 items-center justify-center rounded-md text-muted-foreground transition hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
					aria-label={t("chat.edit_message")}
					title={t("chat.edit_message")}
				>
					<Pencil className="size-3.5" />
				</button>
			) : (
				<button
					type="button"
					onClick={onReply}
					className="flex size-7 items-center justify-center rounded-md text-muted-foreground transition hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
					aria-label={t("chat.reply_action")}
					title={t("chat.reply_action")}
				>
					<Reply className="size-4" />
				</button>
			)}

			<MessageMoreMenuButton {...menuProps} />
		</div>
	);
}
