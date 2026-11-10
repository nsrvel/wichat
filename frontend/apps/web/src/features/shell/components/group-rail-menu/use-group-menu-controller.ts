"use client";

import type { Group } from "@wichat/contracts";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
	canManageGroup,
	isGroupMember,
} from "@/features/shell/lib/group-permissions";
import { translateError } from "@/integrations/i18n/translate-error";
import { APP_HOME_PATH, channelPath } from "@/lib/app-routes";
import { clearLastChannelIfGroup } from "@/lib/last-channel-storage";
import { useLeaveGroupMutation } from "@/services/group/group.mutation";
import { myGroupsQueryOpts } from "@/services/group/group.query";

type UseGroupMenuControllerParams = {
	group: Group;
	activeGroupId?: string;
};

export function useGroupMenuController({ group, activeGroupId }: UseGroupMenuControllerParams) {
	const { t } = useTranslation();
	const router = useRouter();
	const queryClient = useQueryClient();
	const leaveGroup = useLeaveGroupMutation();
	const [leaveDialogOpen, setLeaveDialogOpen] = useState(false);
	const [hideMutedChannels, setHideMutedChannels] = useState(false);
	const [showAllChannels, setShowAllChannels] = useState(false);

	const manage = canManageGroup(group.role);
	const showMemberLeave = isGroupMember(group.role);

	const resetMenuState = useCallback(() => {
		setLeaveDialogOpen(false);
		setHideMutedChannels(false);
		setShowAllChannels(false);
	}, []);

	const openLeaveDialog = useCallback(() => {
		setLeaveDialogOpen(true);
	}, []);

	const showComingSoon = useCallback(() => {
		toast.message(t("shell.group_rail_menu.coming_soon"));
	}, [t]);

	const afterRemovedFromGroup = useCallback(async () => {
		clearLastChannelIfGroup(group.id);
		const remaining = await queryClient.fetchQuery(myGroupsQueryOpts());
		if (activeGroupId === group.id) {
			const next = remaining[0];
			if (next) {
				router.push(channelPath(next.id, next.default_channel_id));
			} else {
				router.push(APP_HOME_PATH);
			}
		}
	}, [activeGroupId, group.id, queryClient, router]);

	const confirmLeave = useCallback(() => {
		leaveGroup.mutate(group.id, {
			onSuccess: () => {
				setLeaveDialogOpen(false);
				void afterRemovedFromGroup();
			},
			onError: (error) => {
				toast.error(translateError(t, error));
			},
		});
	}, [afterRemovedFromGroup, group.id, leaveGroup, t]);

	const copyGroupId = useCallback(
		async (closeMenu: () => void) => {
			try {
				await navigator.clipboard.writeText(group.id);
				toast.success(t("shell.group_rail_menu.copy_id_success"));
				closeMenu();
			} catch {
				toast.error(t("shell.group_rail_menu.coming_soon"));
			}
		},
		[group.id, t],
	);

	return {
		group,
		manage,
		showMemberLeave,
		leaveDialogOpen,
		setLeaveDialogOpen,
		openLeaveDialog,
		confirmLeave,
		hideMutedChannels,
		setHideMutedChannels,
		showAllChannels,
		setShowAllChannels,
		resetMenuState,
		showComingSoon,
		copyGroupId,
		leavePending: leaveGroup.isPending,
	};
}
