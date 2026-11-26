"use client";

import { useChannelNavigation } from "@/features/channel/channel-navigation-provider";
import { QuickViewRoutePage } from "@/features/channel/quick-view-route-page";
import type { QuickViewId } from "@/lib/app-routes";

type Props = {
	groupId: string;
	quickView: QuickViewId;
};

export function QuickViewRoutePageClient({ groupId, quickView }: Props) {
	const nav = useChannelNavigation();

	if (nav.channelId && !nav.quickView) {
		return null;
	}

	const effectiveGroupId = nav.groupId ?? groupId;
	const effectiveQuickView = nav.quickView ?? quickView;

	return <QuickViewRoutePage groupId={effectiveGroupId} quickView={effectiveQuickView} />;
}
