"use client";

import { isPublicEntityId } from "@wichat/contracts";
import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";
import { APP_HOME_PATH, channelPath } from "@/lib/app-routes";
import * as channelQuery from "@/services/channel/channel.query";
import { useQuery } from "@tanstack/react-query";

const LEGACY_GENERAL_SLUG = "general";

type Props = {
	groupId: string;
	channelId: string;
	children: ReactNode;
};

/** Redirect-only guard — does not block the shell on group/channel fetches. */
export function ChannelRouteGuard({ groupId, channelId, children }: Props) {
	const router = useRouter();
	const myGroups = useMyGroupsQuery();
	const channels = useQuery(channelQuery.groupChannelsQueryOpts(groupId));

	const group = myGroups.data?.find((g) => g.id === groupId);
	const defaultChannelId = group?.default_channel_id;

	useEffect(() => {
		if (!isPublicEntityId(groupId)) {
			router.replace(APP_HOME_PATH);
			return;
		}
		if (channelId === LEGACY_GENERAL_SLUG && defaultChannelId) {
			router.replace(channelPath(groupId, defaultChannelId));
			return;
		}
		if (!isPublicEntityId(channelId)) {
			if (defaultChannelId) {
				router.replace(channelPath(groupId, defaultChannelId));
			} else if (!myGroups.isPending) {
				router.replace(APP_HOME_PATH);
			}
			return;
		}
		if (channels.isPending || channels.isError) {
			return;
		}
		const list = channels.data?.channels ?? [];
		if (list.length === 0) {
			return;
		}
		if (!list.some((c) => c.id === channelId)) {
			const fallback = defaultChannelId ?? list[0]?.id;
			if (fallback) {
				router.replace(channelPath(groupId, fallback));
			}
		}
	}, [
		channelId,
		channels.data,
		channels.isError,
		channels.isPending,
		defaultChannelId,
		groupId,
		myGroups.isPending,
		router,
	]);

	return children;
}
