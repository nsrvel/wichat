"use client";

import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { ChatComposer } from "@/features/chat/components/chat-composer";
import { ChannelEmptyState } from "@/features/chat/components/channel-empty-state";
import { ConversationHeader } from "@/features/chat/components/conversation-header";
import { MessageBody } from "@/features/chat/components/message-body";
import { MessageList } from "@/features/chat/components/message-list";
import { MessageReactions } from "@/features/chat/components/message-reactions";
import { MessageRow } from "@/features/chat/components/message-row";
import { toChannelShellNavTarget } from "@/features/channel/lib/channel-shell-nav-target";
import { useChannelMessages } from "@/features/chat/lib/use-channel-messages";
import * as channelQuery from "@/services/channel/channel.query";

type ChannelChatViewProps = {
	groupId: string;
	channelId: string;
	onOpenSidebar?: () => void;
};

export function ChannelChatView({ groupId, channelId, onOpenSidebar }: ChannelChatViewProps) {
	const [draft, setDraft] = useState("");
	const channels = useQuery(channelQuery.groupChannelsQueryOpts(groupId));
	const active = toChannelShellNavTarget(channelId, channels.data?.channels);
	const { messages, isPending } = useChannelMessages(groupId, channelId);

	const showEmpty = !isPending && messages.length === 0;

	return (
		<div className="flex h-full min-w-0 flex-1 flex-col bg-chat-surface text-foreground">
			<div className="lg:hidden">
				<ConversationHeader active={active} onOpenSidebar={onOpenSidebar} />
			</div>

			<div className="flex min-h-0 flex-1 flex-col">
				<MessageList innerClassName={showEmpty ? "min-h-full justify-center py-0" : undefined}>
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
				</MessageList>

				<ChatComposer
					value={draft}
					onChange={setDraft}
					submitDisabled
					onSubmit={() => {
						setDraft("");
					}}
				/>
			</div>
		</div>
	);
}
