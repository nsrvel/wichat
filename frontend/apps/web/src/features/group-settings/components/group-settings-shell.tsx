"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, useEffectEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useGroupSettings } from "@/features/group-settings/group-settings-provider";
import { quickViewPath } from "@/lib/app-routes";
import { hasAnyGroupPermission, hasGroupPermission } from "@/lib/group-permissions";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import { FullscreenSettingsShell } from "@/components/layout/fullscreen-settings/fullscreen-settings-shell";
import { GroupSettingsMainSection } from "@/features/group-settings/components/group-settings-main-section";
import { GroupSettingsSidebar } from "@/features/group-settings/components/group-settings-sidebar";
import type {
	GroupSettingsSection,
	GroupSettingsTarget,
} from "@/features/group-settings/lib/group-settings-target";
import { useGroupSettingsUnsaved } from "@/features/group-settings/lib/group-settings-unsaved-guard";

const GROUP_SETTINGS_TITLE_ID = "group-settings-dialog-title";

type GroupSettingsShellProps = {
	target: GroupSettingsTarget;
	section: GroupSettingsSection;
	onClose: () => void;
};

export function GroupSettingsShell({ target, section, onClose }: GroupSettingsShellProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const { closeGroupSettings } = useGroupSettings();
	const { runIfAllowed } = useGroupSettingsUnsaved();
	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());
	const currentGroup = myGroups.find((g) => g.id === target.groupId);
	const isOwner = target.role === "owner" || currentGroup?.role === "owner";
	const canAccessMembers =
		isOwner || hasAnyGroupPermission(currentGroup, ["kick_members", "ban_members", "manage_roles"]);
	const canAccessBans = isOwner || hasGroupPermission(currentGroup, "ban_members");
	const canAccessAuditLogs = isOwner || hasGroupPermission(currentGroup, "view_audit_log");

	const syncKey = `${target.groupId}:${section}`;
	const baseSection: GroupSettingsSection =
		(section === "members" && !canAccessMembers) ||
		(section === "bans" && !canAccessBans) ||
		(section === "audit_logs" && !canAccessAuditLogs)
			? "profile"
			: section;
	const [sectionOverride, setSectionOverride] = useState<{
		key: string;
		section: GroupSettingsSection;
	} | null>(null);
	const activeSection = sectionOverride?.key === syncKey ? sectionOverride.section : baseSection;

	const openMembersWorkspace = () => {
		runIfAllowed(() => {
			closeGroupSettings();
			router.push(quickViewPath(target.groupId, "members"));
		});
	};

	const triggerMembersWorkspace = useEffectEvent(openMembersWorkspace);

	useEffect(() => {
		if (section !== "members" || !canAccessMembers) {
			return;
		}
		triggerMembersWorkspace();
	}, [section, canAccessMembers, target.groupId]);

	const onSectionChange = (next: GroupSettingsSection) => {
		if (next === activeSection) {
			return;
		}
		if (next === "members" && canAccessMembers) {
			openMembersWorkspace();
			return;
		}
		runIfAllowed(() => setSectionOverride({ key: syncKey, section: next }));
	};

	return (
		<FullscreenSettingsShell
			titleId={GROUP_SETTINGS_TITLE_ID}
			srTitle={`${t("group_settings.header_title")} — ${target.groupName}`}
			onClose={onClose}
			sidebar={
				<GroupSettingsSidebar
					target={target}
					activeSection={activeSection}
					onSectionChange={onSectionChange}
					onOpenMembersInWorkspace={openMembersWorkspace}
				/>
			}
		>
			<GroupSettingsMainSection section={activeSection} target={target} />
		</FullscreenSettingsShell>
	);
}
