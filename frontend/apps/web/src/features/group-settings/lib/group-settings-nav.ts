import type { GroupSettingsSection } from "@/features/group-settings/lib/group-settings-target";

export type GroupSettingsNavItem = {
	id: GroupSettingsSection;
	sectionKey: string;
	section: GroupSettingsSection;
};

export type GroupSettingsNavGroup = {
	labelKey: string;
	items: GroupSettingsNavItem[];
};

export const groupSettingsNavGroups: GroupSettingsNavGroup[] = [
	{
		labelKey: "group_settings.nav.general",
		items: [
			{ id: "profile", sectionKey: "group_settings.nav.profile", section: "profile" },
			{
				id: "onboarding",
				sectionKey: "group_settings.nav.onboarding",
				section: "onboarding",
			},
		],
	},
	{
		labelKey: "group_settings.nav.people",
		items: [
			{ id: "members", sectionKey: "group_settings.nav.members", section: "members" },
			{ id: "roles", sectionKey: "group_settings.nav.roles", section: "roles" },
		],
	},
	{
		labelKey: "group_settings.nav.moderation",
		items: [
			{
				id: "audit_logs",
				sectionKey: "group_settings.nav.audit_logs",
				section: "audit_logs",
			},
			{ id: "bans", sectionKey: "group_settings.nav.bans", section: "bans" },
		],
	},
];
