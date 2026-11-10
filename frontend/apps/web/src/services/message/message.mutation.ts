"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { CreateMessageBody } from "@wichat/contracts";
import { Q_CHANNEL_MESSAGES } from "@/services/message/message.key";
import { createChannelMessage } from "@/services/message/message.service";

type CreateMessageInput = CreateMessageBody & {
	groupId: string;
	channelId: string;
};

export function useCreateMessageMutation() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({ groupId, channelId, body }: CreateMessageInput) =>
			createChannelMessage(groupId, channelId, { body }),
		onSuccess: (_data, variables) => {
			void queryClient.invalidateQueries({
				queryKey: [Q_CHANNEL_MESSAGES, variables.groupId, variables.channelId],
			});
		},
	});
}
