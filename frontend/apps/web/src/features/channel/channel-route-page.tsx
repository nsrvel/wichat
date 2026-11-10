"use client";

import { useEffect } from "react";
import { ChannelChatView } from "@/features/chat/components/channel-chat-view";
import { writeLastChannel } from "@/lib/last-channel-storage";

type Props = {
	groupId: string;
	channelId: string;
};

export function ChannelRoutePage({ groupId, channelId }: Props) {
	useEffect(() => {
		writeLastChannel({ groupId, channelId });
	}, [channelId, groupId]);

	return <ChannelChatView groupId={groupId} channelId={channelId} />;
}
