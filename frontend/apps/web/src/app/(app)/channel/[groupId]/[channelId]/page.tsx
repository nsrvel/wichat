import { ChannelRoutePageClient } from "@/features/channel/channel-route-page-client";

type PageProps = {
	params: Promise<{ groupId: string; channelId: string }>;
};

export default async function ChannelPage({ params }: PageProps) {
	const { groupId, channelId } = await params;
	return <ChannelRoutePageClient groupId={groupId} channelId={channelId} />;
}
