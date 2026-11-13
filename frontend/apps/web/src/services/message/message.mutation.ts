"use client";

import { useMutation, useQueryClient, type InfiniteData } from "@tanstack/react-query";
import type { CreateMessageBody, Message, MessagesList } from "@wichat/contracts";
import { appendOrUpdateInfiniteChannelMessage } from "@/features/chat/lib/channel-message";
import { Q_CHANNEL_MESSAGES } from "@/services/message/message.key";
import { createChannelMessage } from "@/services/message/message.service";

type CreateMessageInput = CreateMessageBody & {
	groupId: string;
	channelId: string;
	authorDisplayName?: string;
	authorId?: string;
};

export function useCreateMessageMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({ groupId, channelId, body }: CreateMessageInput) =>
			createChannelMessage(groupId, channelId, { body }),
		onMutate: async (variables) => {
			const queryKey = [Q_CHANNEL_MESSAGES, variables.groupId, variables.channelId];
			await queryClient.cancelQueries({ queryKey });

			const previousMessages = queryClient.getQueryData<InfiniteData<MessagesList>>(queryKey);

			const tempId = `opt_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
			const optimisticMessage: Message = {
				id: tempId,
				author_id: variables.authorId ?? "me",
				author_display_name: variables.authorDisplayName ?? "You",
				body: variables.body,
				created_at: new Date().toISOString(),
				mine: true,
			};

			queryClient.setQueryData<InfiniteData<MessagesList>>(queryKey, (prev) =>
				appendOrUpdateInfiniteChannelMessage(prev, optimisticMessage, variables.authorId),
			);

			return { previousMessages, queryKey };
		},
		onError: (_err, _variables, context) => {
			if (context?.queryKey && context.previousMessages) {
				queryClient.setQueryData(context.queryKey, context.previousMessages);
			}
		},
		onSuccess: (data, variables) => {
			if (data?.message) {
				queryClient.setQueryData<InfiniteData<MessagesList>>(
					[Q_CHANNEL_MESSAGES, variables.groupId, variables.channelId],
					(prev) => appendOrUpdateInfiniteChannelMessage(prev, data.message, variables.authorId),
				);
			}
		},
	});
}
