"use client";

import { useQuery } from "@tanstack/react-query";
import { Pin, PinOff, X } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { MessageBody } from "@/features/chat/components/message-body";
import { usePinMessageMutation } from "@/services/message/message.mutation";
import { pinnedChannelMessagesQueryOpts } from "@/services/message/message.query";

type PinnedMessagesPopoverProps = {
	groupId: string;
	channelId: string;
	canPin: boolean;
};

export function PinnedMessagesPopover({ groupId, channelId, canPin }: PinnedMessagesPopoverProps) {
	const { t } = useTranslation();
	const [open, setOpen] = useState(false);
	const pinnedQuery = useQuery({
		...pinnedChannelMessagesQueryOpts(groupId, channelId),
		enabled: open && Boolean(groupId) && Boolean(channelId),
	});
	const pinMutation = usePinMessageMutation();
	const pinned = pinnedQuery.data?.messages ?? [];

	return (
		<Popover open={open} onOpenChange={setOpen}>
			<PopoverTrigger
				className="inline-flex h-8 items-center gap-1.5 rounded-md border border-border/70 bg-card/70 px-2.5 text-xs font-medium text-muted-foreground transition hover:bg-accent hover:text-foreground"
				aria-label={t("chat.pinned_messages")}
			>
				<Pin className="size-3.5" />
				<span className="hidden sm:inline">{t("chat.pinned_messages")}</span>
			</PopoverTrigger>
			<PopoverContent align="end" className="w-80 p-0 sm:w-96">
				<div className="flex items-center justify-between border-b border-border px-3.5 py-2.5">
					<div className="flex items-center gap-1.5 text-xs font-semibold text-foreground">
						<Pin className="size-3.5 text-primary" />
						{t("chat.pinned_messages")}
					</div>
					<button
						type="button"
						onClick={() => setOpen(false)}
						className="rounded p-0.5 text-muted-foreground hover:bg-accent hover:text-foreground"
					>
						<X className="size-3.5" />
					</button>
				</div>
				<div className="thin-scroll max-h-80 overflow-y-auto p-2">
					{pinnedQuery.isPending ? (
						<div className="py-6 text-center text-xs text-muted-foreground">…</div>
					) : pinned.length === 0 ? (
						<p className="py-6 text-center text-xs text-muted-foreground">
							{t("chat.no_pinned_messages")}
						</p>
					) : (
						<div className="flex flex-col gap-2">
							{pinned.map((msg) => (
								<div
									key={msg.id}
									className="rounded-lg border border-border/70 bg-muted/30 p-2.5 text-xs"
								>
									<div className="mb-1 flex items-center justify-between gap-2">
										<span className="font-semibold text-foreground">{msg.author_display_name}</span>
										{canPin ? (
											<button
												type="button"
												onClick={() =>
													pinMutation.mutate({
														groupId,
														channelId,
														messageId: msg.id,
														pinned: false,
													})
												}
												className="rounded p-1 text-muted-foreground transition hover:bg-accent hover:text-foreground"
												title={t("chat.unpin_message")}
											>
												<PinOff className="size-3.5" />
											</button>
										) : null}
									</div>
									<MessageBody text={msg.body} className="text-xs leading-snug" />
								</div>
							))}
						</div>
					)}
				</div>
			</PopoverContent>
		</Popover>
	);
}
