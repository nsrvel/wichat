"use client";

import { useQuery, useQueryClient, type InfiniteData } from "@tanstack/react-query";
import { useEffect } from "react";
import { isPublicEntityId } from "@wichat/contracts/common";
import type { MessagesList } from "@wichat/contracts/chat/message";
import {
	applyMessageToInfinite,
	GAP_FILL_BURST_MESSAGES,
	highestSeqInInfinite,
	mergeGapFillIntoInfinite,
} from "@/features/chat/lib/message-sync";
import { removeInfiniteChannelMessage, updateInfiniteChannelMessageReactions } from "@/features/chat/lib/channel-message";
import { parseChannelSeq } from "@/features/chat/lib/message-seq";
import { subscribeChatChannel } from "@/integrations/realtime/chat-channel";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import { Q_CHANNEL_MESSAGES, Q_PINNED_MESSAGES } from "@/services/message/message.key";
import { GAP_FILL_PAGE_SIZE } from "@/features/chat/lib/message-sync";
import { listChannelMessages } from "@/services/message/message.service";

async function replayGapAfterSeq(
	groupId: string,
	channelId: string,
	afterSeq: number,
): Promise<MessagesList> {
	const merged: MessagesList = { messages: [] };
	let cursor = afterSeq;
	let fetched = 0;

	while (fetched < GAP_FILL_BURST_MESSAGES) {
		const page = await listChannelMessages(groupId, channelId, {
			limit: GAP_FILL_PAGE_SIZE,
			afterChannelSeq: cursor > 0 ? String(cursor) : undefined,
		});
		if (!page.messages.length) {
			break;
		}
		merged.messages.push(...page.messages);
		fetched += page.messages.length;
		const last = page.messages[page.messages.length - 1];
		const lastSeq = parseChannelSeq(last.channel_seq);
		if (lastSeq <= cursor) {
			break;
		}
		cursor = lastSeq;
		if (!page.has_more) {
			break;
		}
	}

	return merged;
}

export function useChannelRealtime(
	groupId: string,
	channelId: string,
	onMessageReceived?: (authorId: string) => void,
) {
	const queryClient = useQueryClient();
	const { data: sessionUser } = useQuery(sessionQueryOpts());

	useEffect(() => {
		if (!isPublicEntityId(groupId) || !isPublicEntityId(channelId)) {
			return;
		}

		const queryKey = [Q_CHANNEL_MESSAGES, groupId, channelId];

		return subscribeChatChannel({
			groupId,
			channelId,
			onMessageCreated: (message) => {
				queryClient.setQueryData<InfiniteData<MessagesList>>(queryKey, (prev) =>
					applyMessageToInfinite(prev, message, sessionUser?.id),
				);
				onMessageReceived?.(message.author_id);
			},
			onMessageUpdated: (message) => {
				queryClient.setQueryData<InfiniteData<MessagesList>>(queryKey, (prev) =>
					applyMessageToInfinite(prev, message, sessionUser?.id),
				);
				void queryClient.invalidateQueries({
					queryKey: [Q_PINNED_MESSAGES, groupId, channelId],
				});
			},
			onMessageDeleted: (messageId) => {
				queryClient.setQueryData<InfiniteData<MessagesList>>(queryKey, (prev) =>
					removeInfiniteChannelMessage(prev, messageId),
				);
				void queryClient.invalidateQueries({
					queryKey: [Q_PINNED_MESSAGES, groupId, channelId],
				});
			},
			onMessageReactionsUpdated: (messageId, reactions) => {
				queryClient.setQueryData<InfiniteData<MessagesList>>(queryKey, (prev) =>
					updateInfiniteChannelMessageReactions(prev, messageId, reactions),
				);
			},
			onReconnect: () => {
				const prev = queryClient.getQueryData<InfiniteData<MessagesList>>(queryKey);
				const afterSeq = highestSeqInInfinite(prev);
				if (afterSeq === 0) {
					void queryClient.invalidateQueries({ queryKey });
					return;
				}
				void replayGapAfterSeq(groupId, channelId, afterSeq)
					.then((gapPage) => {
						if (!gapPage.messages.length) {
							return;
						}
						queryClient.setQueryData<InfiniteData<MessagesList>>(queryKey, (current) =>
							mergeGapFillIntoInfinite(current, gapPage, sessionUser?.id),
						);
					})
					.catch(() => {
						// keep cache until manual refresh
					});
			},
		});
	}, [groupId, channelId, queryClient, sessionUser?.id, onMessageReceived]);

	useEffect(() => {
		if (!isPublicEntityId(groupId) || !isPublicEntityId(channelId)) {
			return;
		}
		const queryKey = [Q_CHANNEL_MESSAGES, groupId, channelId];
		return () => {
			queryClient.removeQueries({ queryKey });
			queryClient.removeQueries({ queryKey: [Q_PINNED_MESSAGES, groupId, channelId] });
		};
	}, [groupId, channelId, queryClient]);
}