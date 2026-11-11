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
		items: [{ id: "profile", sectionKey: "group_settings.nav.profile", section: "profile" }],
	},
	{
		labelKey: "group_settings.nav.people",
		items: [
			{ id: "members", sectionKey: "group_settings.nav.members", section: "members" },
			{ id: "roles", sectionKey: "group_settings.nav.roles", section: "roles" },
		],
	},
];
