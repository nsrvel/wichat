import { ChannelRouteGuard } from "@/features/channel/channel-route-guard";

type LayoutProps = {
	children: React.ReactNode;
	params: Promise<{ groupId: string; channelId: string }>;
};

export default async function ChannelRouteLayout({ children, params }: LayoutProps) {
	const { groupId, channelId } = await params;

	return (
		<ChannelRouteGuard groupId={groupId} channelId={channelId}>
			{children}
		</ChannelRouteGuard>
	);
}
