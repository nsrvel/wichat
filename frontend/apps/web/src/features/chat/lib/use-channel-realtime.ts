"use client";

import { useQuery, useQueryClient, type InfiniteData } from "@tanstack/react-query";
import { useEffect } from "react";
import { isPublicEntityId } from "@wichat/contracts/common";
import type { MessagesList } from "@wichat/contracts/chat/message";
import {
	appendOrUpdateInfiniteChannelMessage,
	mergeLatestPageIntoInfiniteMessages,
	removeInfiniteChannelMessage,
	updateInfiniteChannelMessageReactions,
} from "@/features/chat/lib/channel-message";
import { subscribeChatChannel } from "@/integrations/realtime/chat-channel";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import { Q_CHANNEL_MESSAGES, Q_PINNED_MESSAGES } from "@/services/message/message.key";
import {
	CHANNEL_MESSAGES_PAGE_SIZE,
	listChannelMessages,
} from "@/services/message/message.service";

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

		return subscribeChatChannel({
			groupId,
			channelId,
			onMessageCreated: (message) => {
				queryClient.setQueryData<InfiniteData<MessagesList>>(
					[Q_CHANNEL_MESSAGES, groupId, channelId],
					(prev) => appendOrUpdateInfiniteChannelMessage(prev, message, sessionUser?.id),
				);
				onMessageReceived?.(message.author_id);
			},
			onMessageUpdated: (message) => {
				queryClient.setQueryData<InfiniteData<MessagesList>>(
					[Q_CHANNEL_MESSAGES, groupId, channelId],
					(prev) => appendOrUpdateInfiniteChannelMessage(prev, message, sessionUser?.id),
				);
				void queryClient.invalidateQueries({
					queryKey: [Q_PINNED_MESSAGES, groupId, channelId],
				});
			},
			onMessageDeleted: (messageId) => {
				queryClient.setQueryData<InfiniteData<MessagesList>>(
					[Q_CHANNEL_MESSAGES, groupId, channelId],
					(prev) => removeInfiniteChannelMessage(prev, messageId),
				);
				void queryClient.invalidateQueries({
					queryKey: [Q_PINNED_MESSAGES, groupId, channelId],
				});
			},
			onMessageReactionsUpdated: (messageId, reactions) => {
				queryClient.setQueryData<InfiniteData<MessagesList>>(
					[Q_CHANNEL_MESSAGES, groupId, channelId],
					(prev) => updateInfiniteChannelMessageReactions(prev, messageId, reactions),
				);
			},
			onReconnect: () => {
				void listChannelMessages(groupId, channelId, {
					limit: CHANNEL_MESSAGES_PAGE_SIZE,
				})
					.then((latestPage) => {
						queryClient.setQueryData<InfiniteData<MessagesList>>(
							[Q_CHANNEL_MESSAGES, groupId, channelId],
							(prev) => mergeLatestPageIntoInfiniteMessages(prev, latestPage),
						);
					})
					.catch(() => {
						// Fallback: leave cached messages intact until next reconnect/manual retry.
					});
			},
		});
	}, [groupId, channelId, queryClient, sessionUser?.id, onMessageReceived]);
}
