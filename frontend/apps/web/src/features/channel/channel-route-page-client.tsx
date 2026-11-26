"use client";

import { useChannelNavigation } from "@/features/channel/channel-navigation-provider";
import { ChannelRoutePage } from "@/features/channel/channel-route-page";

type Props = {
	groupId: string;
	channelId: string;
};

/**
 * Chat pane keyed off effective navigation (optimistic channel) while the URL catches up.
 */
export function ChannelRoutePageClient({ groupId, channelId }: Props) {
	const nav = useChannelNavigation();
	const effectiveGroupId = nav.groupId ?? groupId;
	const effectiveChannelId = nav.channelId ?? channelId;

	if (nav.quickView) {
		return null;
	}

	return <ChannelRoutePage groupId={effectiveGroupId} channelId={effectiveChannelId} />;
}
