"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { GroupSettingsMembersSection } from "@/features/group-settings/components/members/group-settings-members-section";
import type { GroupSettingsTarget } from "@/features/group-settings/lib/group-settings-target";
import { APP_HOME_PATH, channelPath } from "@/lib/app-routes";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";

type GroupMembersWorkspacePageProps = {
	groupId: string;
};

export function GroupMembersWorkspacePage({ groupId }: GroupMembersWorkspacePageProps) {
	const router = useRouter();
	const myGroups = useMyGroupsQuery();
	const group = myGroups.data?.find((g) => g.id === groupId);
	const isOwner = group?.role === "owner";

	useEffect(() => {
		if (myGroups.isPending) {
			return;
		}
		if (!group) {
			return;
		}
		if (!isOwner) {
			const fallback = group.default_channel_id?.trim();
			router.replace(fallback ? channelPath(groupId, fallback) : APP_HOME_PATH);
		}
	}, [group, groupId, isOwner, myGroups.isPending, router]);

	if (myGroups.isPending || !group || !isOwner) {
		return null;
	}

	const target: GroupSettingsTarget = {
		groupId: group.id,
		groupName: group.name,
		logoUrl: group.logo_url,
		description: group.description,
		bannerUrl: group.banner_url,
		bannerColor: group.banner_color,
		isPrivate: group.is_private,
		role: group.role,
	};

	return (
		<div className="flex h-full min-w-0 flex-1 flex-col bg-chat-surface text-foreground">
			<GroupSettingsMembersSection target={target} layout="workspace" />
		</div>
	);
}
