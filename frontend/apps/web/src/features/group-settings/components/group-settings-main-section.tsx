"use client";

import { GroupSettingsAuditLogsSection } from "@/features/group-settings/components/audit-logs/group-settings-audit-logs-section";
import { GroupSettingsBansSection } from "@/features/group-settings/components/bans/group-settings-bans-section";
import { GroupSettingsPlaceholderSection } from "@/features/group-settings/components/group-settings-placeholder-section";
import { GroupSettingsProfileSection } from "@/features/group-settings/components/group-settings-profile-section";
import { GroupSettingsRolesSection } from "@/features/group-settings/components/roles/group-settings-roles-section";
import {
	groupSettingsMainInnerClass,
	groupSettingsMainScrollClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import { groupSettingsNavGroups } from "@/features/group-settings/lib/group-settings-nav";
import type {
	GroupSettingsSection,
	GroupSettingsTarget,
} from "@/features/group-settings/lib/group-settings-target";

function sectionTitleKey(section: GroupSettingsSection): string {
	for (const group of groupSettingsNavGroups) {
		const item = group.items.find((i) => i.section === section);
		if (item) {
			return item.sectionKey;
		}
	}
	return "group_settings.nav.profile";
}

type GroupSettingsMainSectionProps = {
	section: GroupSettingsSection;
	target: GroupSettingsTarget;
};

export function GroupSettingsMainSection({ section, target }: GroupSettingsMainSectionProps) {
	if (section === "profile") {
		return (
			<div className="flex min-h-0 flex-1 flex-col">
				<GroupSettingsProfileSection target={target} />
			</div>
		);
	}

	if (section === "roles") {
		return (
			<div className="flex min-h-0 flex-1 flex-col">
				<GroupSettingsRolesSection target={target} />
			</div>
		);
	}

	if (section === "bans") {
		return (
			<div className="flex min-h-0 flex-1 flex-col">
				<GroupSettingsBansSection target={target} />
			</div>
		);
	}

	if (section === "audit_logs") {
		return (
			<div className="flex min-h-0 flex-1 flex-col">
				<GroupSettingsAuditLogsSection target={target} />
			</div>
		);
	}

	return (
		<div className={groupSettingsMainScrollClass}>
			<div className={groupSettingsMainInnerClass}>
				<GroupSettingsPlaceholderSection titleKey={sectionTitleKey(section)} />
			</div>
		</div>
	);
}
