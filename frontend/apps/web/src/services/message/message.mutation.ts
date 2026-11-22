"use client";

import { useMutation, useQueryClient, type InfiniteData } from "@tanstack/react-query";
import type {
	CreateMessageBody,
	Message,
	MessageReplyPreview,
	MessagesList,
} from "@wichat/contracts/chat/message";
import { removeInfiniteChannelMessage, updateInfiniteChannelMessageReactions } from "@/features/chat/lib/channel-message";
import { applyMessageToInfinite, replacePendingMessage } from "@/features/chat/lib/message-sync";
import { newClientMessageId } from "@/features/chat/lib/message-seq";
import { Q_CHANNEL_MESSAGES, Q_PINNED_MESSAGES } from "@/services/message/message.key";
import {
	createChannelMessage,
	deleteChannelMessage,
	pinChannelMessage,
	toggleMessageReaction,
	unpinChannelMessage,
	updateChannelMessage,
} from "@/services/message/message.service";

type CreateMessageInput = CreateMessageBody & {
	groupId: string;
	channelId: string;
	authorDisplayName?: string;
	authorAvatarUrl?: string;
	authorId?: string;
	replyPreview?: MessageReplyPreview;
};

export function useCreateMessageMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({ groupId, channelId, body, reply_to_id, client_msg_id }: CreateMessageInput) =>
			createChannelMessage(groupId, channelId, { body, reply_to_id, client_msg_id }),
		onMutate: async (variables) => {
			const queryKey = [Q_CHANNEL_MESSAGES, variables.groupId, variables.channelId];
			await queryClient.cancelQueries({ queryKey });

			const previousMessages = queryClient.getQueryData<InfiniteData<MessagesList>>(queryKey);

			const clientMsgId = variables.client_msg_id ?? newClientMessageId();
			variables.client_msg_id = clientMsgId;
			const pendingId = `pending_${clientMsgId}`;
			const optimisticMessage: Message = {
				id: pendingId,
				client_msg_id: clientMsgId,
				author_id: variables.authorId ?? "me",
				author_display_name: variables.authorDisplayName ?? "You",
				author_avatar_url: variables.authorAvatarUrl ?? "",
				body: variables.body,
				created_at: new Date().toISOString(),
				edited_at: "",
				is_pinned: false,
				reply_to: variables.replyPreview ?? null,
				reactions: [],
				mine: true,
			};

			queryClient.setQueryData<InfiniteData<MessagesList>>(queryKey, (prev) =>
				applyMessageToInfinite(prev, optimisticMessage, variables.authorId),
			);

			return { previousMessages, queryKey, pendingId, clientMsgId };
		},
		onError: (_err, _variables, context) => {
			if (context?.queryKey && context.previousMessages) {
				queryClient.setQueryData(context.queryKey, context.previousMessages);
			}
		},
		onSuccess: (data, variables, context) => {
			if (data?.message) {
				const enriched = {
					...data.message,
					client_msg_id: context?.clientMsgId ?? data.message.client_msg_id,
				};
				queryClient.setQueryData<InfiniteData<MessagesList>>(
					[Q_CHANNEL_MESSAGES, variables.groupId, variables.channelId],
					(prev) =>
						context?.pendingId
							? replacePendingMessage(prev, context.pendingId, enriched, variables.authorId)
							: applyMessageToInfinite(prev, enriched, variables.authorId),
				);
			}
		},
	});
}

export function useUpdateMessageMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({
			groupId,
			channelId,
			messageId,
			body,
		}: {
			groupId: string;
			channelId: string;
			messageId: string;
			body: string;
		}) => updateChannelMessage(groupId, channelId, messageId, { body }),
		onSuccess: (data, variables) => {
			if (data?.message) {
				queryClient.setQueryData<InfiniteData<MessagesList>>(
					[Q_CHANNEL_MESSAGES, variables.groupId, variables.channelId],
					(prev) => applyMessageToInfinite(prev, data.message, data.message.author_id),
				);
			}
		},
	});
}

export function useDeleteMessageMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({
			groupId,
			channelId,
			messageId,
		}: {
			groupId: string;
			channelId: string;
			messageId: string;
		}) => deleteChannelMessage(groupId, channelId, messageId),
		onMutate: async (variables) => {
			const queryKey = [Q_CHANNEL_MESSAGES, variables.groupId, variables.channelId];
			await queryClient.cancelQueries({ queryKey });
			const previousMessages = queryClient.getQueryData<InfiniteData<MessagesList>>(queryKey);
			queryClient.setQueryData<InfiniteData<MessagesList>>(queryKey, (prev) =>
				removeInfiniteChannelMessage(prev, variables.messageId),
			);
			return { previousMessages, queryKey };
		},
		onError: (_err, _variables, context) => {
			if (context?.queryKey && context.previousMessages) {
				queryClient.setQueryData(context.queryKey, context.previousMessages);
			}
		},
		onSuccess: (_data, variables) => {
			void queryClient.invalidateQueries({
				queryKey: [Q_PINNED_MESSAGES, variables.groupId, variables.channelId],
			});
		},
	});
}

export function useToggleReactionMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({
			groupId,
			channelId,
			messageId,
			emoji,
		}: {
			groupId: string;
			channelId: string;
			messageId: string;
			emoji: string;
		}) => toggleMessageReaction(groupId, channelId, messageId, { emoji }),
		onSuccess: (data, variables) => {
			queryClient.setQueryData<InfiniteData<MessagesList>>(
				[Q_CHANNEL_MESSAGES, variables.groupId, variables.channelId],
				(prev) => updateInfiniteChannelMessageReactions(prev, variables.messageId, data.reactions),
			);
		},
	});
}

export function usePinMessageMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({
			groupId,
			channelId,
			messageId,
			pinned,
		}: {
			groupId: string;
			channelId: string;
			messageId: string;
			pinned: boolean;
		}) =>
			pinned
				? pinChannelMessage(groupId, channelId, messageId)
				: unpinChannelMessage(groupId, channelId, messageId),
		onSuccess: (data, variables) => {
			if (data?.message) {
				queryClient.setQueryData<InfiniteData<MessagesList>>(
					[Q_CHANNEL_MESSAGES, variables.groupId, variables.channelId],
					(prev) => applyMessageToInfinite(prev, data.message),
				);
			}
			void queryClient.invalidateQueries({
				queryKey: [Q_PINNED_MESSAGES, variables.groupId, variables.channelId],
			});
		},
	});
}
