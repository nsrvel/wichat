"use client";

import { useEffect, useRef } from "react";
import { ChatComposer } from "@/features/chat/components/chat-composer";
import { ConversationHeader } from "@/features/chat/components/conversation-header";
import { MessageList } from "@/features/chat/components/message-list";
import { DemoChannelThread } from "@/features/dev/chat-ui/components/demo-channel-thread";
import { useShellNavigation } from "@/components/layout/shell-navigation";

export function ChannelView({ onOpenSidebar }: { onOpenSidebar?: () => void }) {
	const { active } = useShellNavigation();
	const scrollRef = useRef<HTMLDivElement>(null);
	useEffect(() => {
		scrollRef.current?.scrollTo({ top: 38 });
	}, []);

	return (
		<div className="flex h-full min-w-0 flex-1 flex-col bg-chat-surface text-foreground">
			<div className="lg:hidden">
				<ConversationHeader active={active} onOpenSidebar={onOpenSidebar} />
			</div>

			<div className="flex min-h-0 flex-1 flex-col">
				<MessageList scrollRef={scrollRef}>
					<DemoChannelThread />
				</MessageList>

				<ChatComposer />
			</div>
		</div>
	);
}
