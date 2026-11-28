"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { dmSidebarPeerMember } from "@/features/channel/lib/dm-sidebar-peer";
import { toChannelShellNavTarget } from "@/features/channel/lib/channel-shell-nav-target";
import { ChannelChatView } from "@/features/chat/components/channel-chat-view";
import { ConversationHeader } from "@/features/chat/components/conversation-header";
import { MentionsListSidebar } from "@/features/mentions/components/mentions-list-sidebar";
import type { MentionsFilterId } from "@/features/mentions/lib/mentions-filters";
import { dispatchJumpToMessage } from "@/features/search/lib/jump-to-message-event";
import { quickViewPath } from "@/lib/app-routes";
import { cn } from "@/lib/utils";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import { groupChannelsQueryOpts } from "@/services/channel/channel.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";

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
	const [filter, setFilter] = useState<MentionsFilterId>("all");

	const { data: sessionUser } = useQuery(sessionQueryOpts());
	const channelsQuery = useQuery({
		...groupChannelsQueryOpts(groupId),
		enabled: Boolean(groupId && channelId),
	});
	const membersQuery = useQuery({
		...groupMembersQueryOpts(groupId, ""),
		enabled: Boolean(groupId && channelId),
	});

	const activeChannelTarget = useMemo(() => {
		if (!channelId) {
			return null;
		}
		const channels = channelsQuery.data?.channels ?? [];
		const channel = channels.find((c) => c.id === channelId);
		const peer =
			channel?.type === "dm"
				? dmSidebarPeerMember(channel, membersQuery.data?.members, sessionUser?.id)
				: undefined;
		return toChannelShellNavTarget(channelId, channels, peer);
	}, [channelId, channelsQuery.data?.channels, membersQuery.data?.members, sessionUser?.id]);

	const handleFilterChange = useCallback((next: MentionsFilterId) => {
		setFilter((prev) => (prev === next ? "all" : next));
	}, []);

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

	const chatPane =
		channelId && activeChannelTarget ? (
			<div className="flex h-full min-h-0 min-w-0 flex-1 flex-col">
				<div className="shrink-0 border-b border-black">
					<ConversationHeader
						active={activeChannelTarget}
						onOpenSidebar={onOpenSidebar}
						showBottomDivider={false}
					/>
				</div>
				<div className="min-h-0 flex-1">
					<ChannelChatView
						groupId={groupId}
						channelId={channelId}
						onOpenSidebar={onOpenSidebar}
						embedInWorkspace
						jumpToMessageId={messageId}
					/>
				</div>
			</div>
		) : (
			<div className="flex flex-1 items-center justify-center px-6 text-center text-sm text-muted-foreground">
				{t("shell.mentions.select_hint")}
			</div>
		);

	return (
		<div className="flex h-full min-w-0 flex-1 flex-col bg-chat-surface text-foreground">
			<div className="flex min-h-0 flex-1">
				<div
					className={cn("min-h-0 shrink-0", channelId ? "hidden lg:flex" : "flex w-full lg:w-auto")}
				>
					<MentionsListSidebar
						groupId={groupId}
						filter={filter}
						selectedChannelId={channelId}
						selectedMessageId={messageId}
						onSelect={updateSelection}
						onFilterChange={handleFilterChange}
					/>
				</div>
				<div className={cn("min-h-0 min-w-0 flex-1", !channelId && "hidden lg:flex")}>
					{chatPane}
				</div>
			</div>
		</div>
	);
}
