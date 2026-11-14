"use client";

import { useTranslation } from "react-i18next";
import {
	USER_SETTINGS_ITEM_BY_ID,
	USER_SETTINGS_PLACEHOLDER_ICONS,
	type UserSettingsSectionId,
} from "@/features/user-settings/lib/user-settings-nav";

export function UserSettingsPlaceholderSection({
	sectionId,
}: {
	sectionId: UserSettingsSectionId;
}) {
	const { t } = useTranslation();
	const item = USER_SETTINGS_ITEM_BY_ID[sectionId];
	const Icon = USER_SETTINGS_PLACEHOLDER_ICONS[sectionId] ?? item.icon;

	return (
		<div className="flex flex-col items-start gap-3 rounded-xl border border-border/70 bg-muted/20 p-6">
			<span className="flex size-10 items-center justify-center rounded-xl bg-muted text-foreground">
				<Icon className="size-5" aria-hidden />
			</span>
			<div>
				<h3 className="text-base font-semibold text-foreground">{t(item.labelKey)}</h3>
				<p className="mt-1 text-sm text-muted-foreground">{t("user_settings.coming_soon_hint")}</p>
			</div>
		</div>
	);
}
