"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, useEffectEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useGroupSettings } from "@/features/group-settings/group-settings-provider";
import { quickViewPath } from "@/lib/app-routes";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import { GroupSettingsCloseButton } from "@/features/group-settings/components/group-settings-close-button";
import { GroupSettingsMainSection } from "@/features/group-settings/components/group-settings-main-section";
import { GroupSettingsSidebar } from "@/features/group-settings/components/group-settings-sidebar";
import { groupSettingsBodyOuterClass, groupSettingsCanvasClass, groupSettingsMainPaneClass, groupSettingsShellClass, groupSettingsSidebarPaneClass, groupSettingsSplitBackdropClass } from "@/features/group-settings/lib/group-settings-chrome";
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
	const isOwner =
		target.role === "owner" || myGroups.some((g) => g.id === target.groupId && g.role === "owner");

	const syncKey = `${target.groupId}:${section}`;
	const baseSection: GroupSettingsSection = section === "members" && !isOwner ? "profile" : section;
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
		if (section !== "members" || !isOwner) {
			return;
		}
		triggerMembersWorkspace();
	}, [section, isOwner, target.groupId]);

	const onSectionChange = (next: GroupSettingsSection) => {
		if (next === activeSection) {
			return;
		}
		if (next === "members" && isOwner) {
			openMembersWorkspace();
			return;
		}
		runIfAllowed(() => setSectionOverride({ key: syncKey, section: next }));
	};

	return (
		<div className={groupSettingsShellClass}>
			<h1 id={GROUP_SETTINGS_TITLE_ID} className="sr-only">
				{t("group_settings.header_title")} — {target.groupName}
			</h1>
			<GroupSettingsCloseButton onClose={onClose} />
			<div className={groupSettingsSplitBackdropClass} aria-hidden />
			<div className={groupSettingsBodyOuterClass}>
				<div className={groupSettingsCanvasClass}>
					<div className={groupSettingsSidebarPaneClass}>
						<GroupSettingsSidebar
							target={target}
							activeSection={activeSection}
							onSectionChange={onSectionChange}
							onOpenMembersInWorkspace={openMembersWorkspace}
						/>
					</div>
					<div className={groupSettingsMainPaneClass} role="tabpanel">
						<GroupSettingsMainSection section={activeSection} target={target} />
					</div>
				</div>
			</div>
		</div>
	);
}
