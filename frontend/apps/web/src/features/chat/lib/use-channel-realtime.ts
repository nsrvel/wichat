"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { isPublicEntityId, type MessagesList } from "@wichat/contracts";
import { subscribeChatChannel } from "@/integrations/realtime/chat-channel";
import { Q_CHANNEL_MESSAGES } from "@/services/message/message.key";

export function useChannelRealtime(groupId: string, channelId: string) {
	const queryClient = useQueryClient();

	useEffect(() => {
		if (!isPublicEntityId(groupId) || !isPublicEntityId(channelId)) {
			return;
		}

		return subscribeChatChannel({
			groupId,
			channelId,
			onMessageCreated: (message) => {
				queryClient.setQueryData<MessagesList>([Q_CHANNEL_MESSAGES, groupId, channelId], (prev) => {
					if (!prev) {
						return { messages: [message] };
					}
					if (prev.messages.some((m) => m.id === message.id)) {
						return prev;
					}
					return { messages: [...prev.messages, message] };
				});
			},
		});
	}, [groupId, channelId, queryClient]);
}
