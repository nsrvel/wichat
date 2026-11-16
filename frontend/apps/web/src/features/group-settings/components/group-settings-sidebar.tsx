"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "sonner";
import { Ban, Compass, FileText, Shield, SquareArrowOutUpRight, User, Users } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useTranslation } from "react-i18next";
import {
	FullscreenSettingsSidebar,
	type FullscreenSettingsNavGroup,
} from "@/components/layout/fullscreen-settings/fullscreen-settings-sidebar";
import { DeleteGroupConfirmDialog } from "@/features/group-settings/components/delete-group-confirm-dialog";
import { useGroupSettings } from "@/features/group-settings/group-settings-provider";
import { groupSettingsNavGroups } from "@/features/group-settings/lib/group-settings-nav";
import type {
	GroupSettingsSection,
	GroupSettingsTarget,
} from "@/features/group-settings/lib/group-settings-target";
import { translateError } from "@/integrations/i18n/translate-error";
import { APP_HOME_PATH, channelPath } from "@/lib/app-routes";
import { hasAnyGroupPermission, hasGroupPermission } from "@/lib/group-permissions";
import { clearLastChannelIfGroup } from "@/lib/last-channel-storage";
import { useDeleteGroupMutation } from "@/services/group/group.mutation";
import { myGroupsQueryOpts } from "@/services/group/group.query";

const NAV_ICON: Record<GroupSettingsSection, LucideIcon> = {
	profile: User,
	onboarding: Compass,
	members: Users,
	roles: Shield,
	audit_logs: FileText,
	bans: Ban,
};

type GroupSettingsSidebarProps = {
	target: GroupSettingsTarget;
	activeSection: GroupSettingsSection;
	onSectionChange: (section: GroupSettingsSection) => void;
	onOpenMembersInWorkspace: () => void;
};

export function GroupSettingsSidebar({
	target,
	activeSection,
	onSectionChange,
	onOpenMembersInWorkspace,
}: GroupSettingsSidebarProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const queryClient = useQueryClient();
	const { closeGroupSettings } = useGroupSettings();
	const deleteGroupMutation = useDeleteGroupMutation();
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());
	const currentGroup = myGroups.find((g) => g.id === target.groupId);
	const isOwner = target.role === "owner" || currentGroup?.role === "owner";
	const canSeeRoles = isOwner || hasGroupPermission(currentGroup, "manage_roles");
	const canSeeMembers =
		isOwner || hasAnyGroupPermission(currentGroup, ["kick_members", "ban_members", "manage_roles"]);
	const canSeeBans = isOwner || hasGroupPermission(currentGroup, "ban_members");
	const canSeeAuditLogs = isOwner || hasGroupPermission(currentGroup, "view_audit_log");

	const handleDeleteConfirm = () => {
		deleteGroupMutation.mutate(target.groupId, {
			onSuccess: async () => {
				setDeleteDialogOpen(false);
				closeGroupSettings();
				clearLastChannelIfGroup(target.groupId);
				const remaining = await queryClient.fetchQuery(myGroupsQueryOpts());
				const next = remaining.find((g) => g.id !== target.groupId);
				if (next) {
					router.push(channelPath(next.id, next.default_channel_id));
				} else {
					router.push(APP_HOME_PATH);
				}
			},
			onError: (error) => {
				toast.error(translateError(t, error));
			},
		});
	};

	const navGroups: FullscreenSettingsNavGroup[] = groupSettingsNavGroups.map((group) => ({
		id: group.labelKey,
		label: t(group.labelKey),
		items: group.items
			.filter((item) => {
				if (item.id === "members") return canSeeMembers;
				if (item.id === "roles") return canSeeRoles;
				if (item.id === "bans") return canSeeBans;
				if (item.id === "audit_logs") return canSeeAuditLogs;
				return true;
			})
			.map((item) => ({
				id: item.id,
				label: t(item.sectionKey),
				icon: NAV_ICON[item.id],
				selected: item.section === activeSection,
				onSelect: () => {
					if (item.id === "members") {
						onOpenMembersInWorkspace();
					} else {
						onSectionChange(item.section);
					}
				},
				trailingIcon:
					item.id === "members" ? (
						<SquareArrowOutUpRight className="size-3.5" strokeWidth={2} aria-hidden />
					) : undefined,
			})),
	}));

	return (
		<>
			{isOwner ? (
				<DeleteGroupConfirmDialog
					groupName={target.groupName}
					open={deleteDialogOpen}
					onOpenChange={setDeleteDialogOpen}
					onConfirm={handleDeleteConfirm}
					confirmDisabled={deleteGroupMutation.isPending}
				/>
			) : null}
			<FullscreenSettingsSidebar
				headerTitle={target.groupName}
				navLabel={t("group_settings.nav_label")}
				groups={navGroups}
				destructiveAction={
					isOwner
						? {
								label: t("shell.delete_group"),
								onClick: () => setDeleteDialogOpen(true),
							}
						: null
				}
			/>
		</>
	);
}
