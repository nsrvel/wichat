"use client";

import { useTranslation } from "react-i18next";
import {
	groupSettingsSectionLeadClass,
	groupSettingsSectionTitleClass,
} from "@/features/group-settings/lib/group-settings-chrome";

type GroupSettingsPlaceholderSectionProps = {
	titleKey: string;
};

export function GroupSettingsPlaceholderSection({
	titleKey,
}: GroupSettingsPlaceholderSectionProps) {
	const { t } = useTranslation();

	return (
		<div className="flex w-full flex-col gap-2">
			<h2 className={groupSettingsSectionTitleClass}>{t(titleKey)}</h2>
			<p className={groupSettingsSectionLeadClass}>{t("shell.group_rail_menu.coming_soon")}</p>
		</div>
	);
}
