"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { ChannelChatView } from "@/features/chat/components/channel-chat-view";
import { MentionsListSidebar } from "@/features/mentions/components/mentions-list-sidebar";
import type { MentionsFilterId } from "@/features/mentions/lib/mentions-filters";
import { dispatchJumpToMessage } from "@/features/search/lib/jump-to-message-event";
import { quickViewPath } from "@/lib/app-routes";
import { cn } from "@/lib/utils";

type MentionsWorkspacePageProps = {
	groupId: string;
	onOpenSidebar?: () => void;
};

export function MentionsWorkspacePage({ groupId, onOpenSidebar }: MentionsWorkspacePageProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const searchParams = useSearchParams();
	const channelId = searchParams.get("channelId")?.trim() ?? "";
	const messageId = searchParams.get("messageId")?.trim() ?? "";
	const [filter, setFilter] = useState<MentionsFilterId>("channels");

	const updateSelection = useCallback(
		(nextChannelId: string, nextMessageId: string) => {
			const params = new URLSearchParams();
			params.set("channelId", nextChannelId);
			params.set("messageId", nextMessageId);
			router.replace(`${quickViewPath(groupId, "mentions")}?${params.toString()}`);
		},
		[groupId, router],
	);

	useEffect(() => {
		if (channelId && messageId) {
			dispatchJumpToMessage(messageId);
		}
	}, [channelId, messageId]);

	const chatPane = channelId ? (
			<ChannelChatView
				groupId={groupId}
				channelId={channelId}
				onOpenSidebar={onOpenSidebar}
				embedInWorkspace
				jumpToMessageId={messageId}
			/>
		) : (
			<div className="flex flex-1 items-center justify-center px-6 text-center text-sm text-muted-foreground">
				{t("shell.mentions.select_hint")}
			</div>
		);

	return (
		<div className="flex h-full min-w-0 flex-1 flex-col bg-chat-surface text-foreground">
			<div className="flex min-h-0 flex-1">
				<div
					className={cn(
						"min-h-0 shrink-0",
						channelId ? "hidden lg:flex" : "flex w-full lg:w-auto",
					)}
				>
					<MentionsListSidebar
						groupId={groupId}
						filter={filter}
						selectedChannelId={channelId}
						selectedMessageId={messageId}
						onSelect={updateSelection}
						onFilterChange={setFilter}
					/>
				</div>
				<div className={cn("min-h-0 min-w-0 flex-1", !channelId && "hidden lg:flex")}>{chatPane}</div>
			</div>
		</div>
	);
}