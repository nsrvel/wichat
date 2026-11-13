"use client";

import { useQuery } from "@tanstack/react-query";
import { ChevronDown } from "lucide-react";
import { useCallback, useLayoutEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useCreateMessageMutation } from "@/services/message/message.mutation";
import { ChatComposer } from "@/features/chat/components/chat-composer";
import { ChannelEmptyState } from "@/features/chat/components/channel-empty-state";
import { ConversationHeader } from "@/features/chat/components/conversation-header";
import { MessageBody } from "@/features/chat/components/message-body";
import { MessageList } from "@/features/chat/components/message-list";
import { MessageReactions } from "@/features/chat/components/message-reactions";
import { MessageRow } from "@/features/chat/components/message-row";
import { toChannelShellNavTarget } from "@/features/channel/lib/channel-shell-nav-target";
import { useChannelMessages } from "@/features/chat/lib/use-channel-messages";
import { useChannelRealtime } from "@/features/chat/lib/use-channel-realtime";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import * as channelQuery from "@/services/channel/channel.query";

type ChannelChatViewProps = {
	groupId: string;
	channelId: string;
	onOpenSidebar?: () => void;
};

export function ChannelChatView({ groupId, channelId, onOpenSidebar }: ChannelChatViewProps) {
	const { t } = useTranslation();
	const scrollRef = useRef<HTMLDivElement>(null);
	const bottomSentinelRef = useRef<HTMLDivElement>(null);
	const anchorRef = useRef<{ scrollHeight: number; scrollTop: number } | null>(null);
	const isNearBottomRef = useRef(true);
	const shouldScrollToBottomRef = useRef(false);
	const [showJumpToLatest, setShowJumpToLatest] = useState(false);

	const channels = useQuery(channelQuery.groupChannelsQueryOpts(groupId));
	const active = toChannelShellNavTarget(channelId, channels.data?.channels);
	const { data: sessionUser } = useQuery(sessionQueryOpts());
	const {
		messages,
		isPending,
		isError,
		hasOlderMessages,
		isFetchingOlderMessages,
		fetchOlderMessages,
		refetch,
	} = useChannelMessages(groupId, channelId);

	const scrollToBottom = useCallback((behavior: ScrollBehavior = "auto") => {
		const el = scrollRef.current;
		if (el) {
			el.scrollTop = el.scrollHeight;
		}
		bottomSentinelRef.current?.scrollIntoView({ behavior, block: "end" });

		// Ensure layout paint has registered new message height
		requestAnimationFrame(() => {
			if (el) {
				el.scrollTop = el.scrollHeight;
			}
			bottomSentinelRef.current?.scrollIntoView({ behavior, block: "end" });
		});
	}, []);

	const handleMessageReceived = useCallback(() => {
		if (!isNearBottomRef.current) {
			setShowJumpToLatest(true);
		}
	}, []);

	useChannelRealtime(groupId, channelId, handleMessageReceived);
	const createMessage = useCreateMessageMutation();

	const showEmpty = !isPending && !isError && messages.length === 0;

	// Scroll anchoring & position restoration
	useLayoutEffect(() => {
		const el = scrollRef.current;
		if (!el) {
			return;
		}

		if (anchorRef.current) {
			const delta = el.scrollHeight - anchorRef.current.scrollHeight;
			el.scrollTop = anchorRef.current.scrollTop + delta;
			anchorRef.current = null;
			return;
		}

		const latestMessage = messages[messages.length - 1];
		const isLatestMine =
			latestMessage?.mine ||
			(Boolean(sessionUser?.id) && latestMessage?.authorId === sessionUser?.id);

		if (shouldScrollToBottomRef.current || isLatestMine || isNearBottomRef.current) {
			shouldScrollToBottomRef.current = false;
			isNearBottomRef.current = true;
			setShowJumpToLatest(false);
			scrollToBottom("auto");
		}
	}, [messages, channelId, sessionUser?.id, scrollToBottom]);

	const handleScroll = useCallback(() => {
		const el = scrollRef.current;
		if (!el) {
			return;
		}

		const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
		const nearBottom = distanceFromBottom <= 120;
		isNearBottomRef.current = nearBottom;
		if (nearBottom) {
			setShowJumpToLatest(false);
		}

		if (el.scrollTop <= 120 && hasOlderMessages && !isFetchingOlderMessages) {
			anchorRef.current = {
				scrollHeight: el.scrollHeight,
				scrollTop: el.scrollTop,
			};
			fetchOlderMessages();
		}
	}, [hasOlderMessages, isFetchingOlderMessages, fetchOlderMessages]);

	const jumpToLatest = useCallback(() => {
		scrollToBottom("smooth");
		setShowJumpToLatest(false);
	}, [scrollToBottom]);

	const handleSendMessage = useCallback(
		(body: string) => {
			shouldScrollToBottomRef.current = true;
			isNearBottomRef.current = true;
			setShowJumpToLatest(false);
			scrollToBottom("auto");

			createMessage.mutate(
				{
					groupId,
					channelId,
					body,
					authorDisplayName: sessionUser?.display_name,
					authorId: sessionUser?.id,
				},
				{
					onSuccess: () => {
						scrollToBottom("smooth");
					},
				},
			);
		},
		[createMessage, groupId, channelId, sessionUser, scrollToBottom],
	);

	return (
		<div className="relative flex h-full min-w-0 flex-1 flex-col bg-chat-surface text-foreground">
			<div className="lg:hidden">
				<ConversationHeader active={active} onOpenSidebar={onOpenSidebar} />
			</div>

			<div className="flex min-h-0 flex-1 flex-col">
				<MessageList
					scrollRef={scrollRef}
					onScroll={handleScroll}
					innerClassName={showEmpty || isError ? "min-h-full justify-center py-0" : undefined}
				>
					{isError ? (
						<div className="flex flex-col items-center justify-center gap-3 py-8 text-center text-sm">
							<p className="text-muted-foreground">{t("chat.load_error")}</p>
							<button
								type="button"
								onClick={refetch}
								className="rounded-md border border-border bg-card px-3 py-1.5 text-xs font-medium text-foreground shadow-xs transition hover:bg-muted"
							>
								{t("chat.retry")}
							</button>
						</div>
					) : null}

					{isFetchingOlderMessages ? (
						<div className="flex items-center justify-center gap-2 py-3 text-xs text-muted-foreground">
							<span className="inline-block size-3.5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
							{t("chat.loading_older")}
						</div>
					) : null}

					{!hasOlderMessages && messages.length > 0 && !isError ? (
						<div className="flex flex-col items-center justify-center gap-1.5 border-b border-border/40 pb-6 pt-8 text-center">
							<div className="flex size-12 items-center justify-center rounded-full bg-muted text-lg font-bold text-foreground">
								#
							</div>
							<h3 className="text-base font-semibold text-foreground">
								{t("chat.history_start_title", {
									name: active?.channelName || active?.label || "channel",
								})}
							</h3>
							<p className="max-w-md text-xs text-muted-foreground">
								{t("chat.history_start_description", {
									name: active?.channelName || active?.label || "channel",
								})}
							</p>
						</div>
					) : null}

					{showEmpty ? <ChannelEmptyState /> : null}

					{messages.map((message) => (
						<MessageRow
							key={message.id}
							authorAvatarUrl={message.authorAvatarUrl}
							authorDisplayName={message.authorDisplayName}
							timeLabel={message.timeLabel}
							mine={message.mine}
							reactions={
								message.reactions && message.reactions.length > 0 ? (
									<MessageReactions reactions={message.reactions} />
								) : undefined
							}
						>
							<MessageBody text={message.bodyPlain} />
						</MessageRow>
					))}

					<div ref={bottomSentinelRef} className="h-px w-full shrink-0" aria-hidden />
				</MessageList>

				{showJumpToLatest ? (
					<div className="absolute bottom-20 left-1/2 z-10 -translate-x-1/2 animate-in fade-in slide-in-from-bottom-2 duration-150">
						<button
							type="button"
							onClick={jumpToLatest}
							className="flex items-center gap-1.5 rounded-full border border-border bg-primary px-3.5 py-1.5 text-xs font-semibold text-primary-foreground shadow-lg transition hover:brightness-110 active:scale-95"
						>
							<ChevronDown className="size-3.5" />
							{t("chat.jump_to_latest")}
						</button>
					</div>
				) : null}

				<ChatComposer onSubmit={handleSendMessage} />
			</div>
		</div>
	);
}
