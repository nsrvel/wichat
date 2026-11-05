import { ChannelRoutePage } from "@/features/channel/channel-route-page";

type PageProps = {
	params: Promise<{ groupId: string; channelId: string }>;
};

export default async function ChannelPage({ params }: PageProps) {
	const { groupId, channelId } = await params;
	return <ChannelRoutePage groupId={groupId} channelId={channelId} />;
}
