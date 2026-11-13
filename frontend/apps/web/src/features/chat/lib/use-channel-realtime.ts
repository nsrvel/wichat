"use client";

import { useQuery, useQueryClient, type InfiniteData } from "@tanstack/react-query";
import { useEffect } from "react";
import { isPublicEntityId, type MessagesList } from "@wichat/contracts";
import { appendOrUpdateInfiniteChannelMessage } from "@/features/chat/lib/channel-message";
import { subscribeChatChannel } from "@/integrations/realtime/chat-channel";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import { Q_CHANNEL_MESSAGES } from "@/services/message/message.key";

export function useChannelRealtime(
	groupId: string,
	channelId: string,
	onMessageReceived?: () => void,
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
				onMessageReceived?.();
			},
		});
	}, [groupId, channelId, queryClient, sessionUser?.id, onMessageReceived]);
}
