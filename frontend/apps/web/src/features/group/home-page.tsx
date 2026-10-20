"use client";

import { useQuery } from "@tanstack/react-query";
import { ChannelView } from "@/features/chat/components/channel-view";
import { GroupChooser } from "@/features/group/group-chooser";
import * as groupQuery from "@/services/group/group.query";

export function HomePage() {
	const myGroups = useQuery(groupQuery.myGroupsQueryOpts());

	if (myGroups.isPending) {
		return <div className="flex-1" suppressHydrationWarning />;
	}

	if (myGroups.isError || !myGroups.data?.length) {
		return <GroupChooser />;
	}

	return <ChannelView />;
}
