import { ChatShellLayout } from "@/features/shell/chat-shell-layout";

type LayoutProps = {
	children: React.ReactNode;
	params: Promise<{ groupId: string; channelId: string }>;
};

export default async function ChannelChatShellLayout({ children, params }: LayoutProps) {
	const { groupId, channelId } = await params;

	return (
		<ChatShellLayout groupId={groupId} channelId={channelId}>
			{children}
		</ChatShellLayout>
	);
}
