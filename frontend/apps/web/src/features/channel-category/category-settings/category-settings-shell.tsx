"use client";

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { FullscreenSettingsShell } from "@/components/layout/fullscreen-settings/fullscreen-settings-shell";
import { CategorySettingsOverviewSection } from "@/features/channel-category/category-settings/category-settings-overview-section";
import { CategorySettingsPermissionsSection } from "@/features/channel-category/category-settings/category-settings-permissions-section";
import {
	CategorySettingsSidebar,
	type CategorySettingsSection,
	type CategorySettingsTarget,
} from "@/features/channel-category/category-settings/category-settings-sidebar";
import { useGroupSettingsUnsaved } from "@/features/group-settings/lib/group-settings-unsaved-guard";

const CATEGORY_SETTINGS_TITLE_ID = "category-settings-dialog-title";

type CategorySettingsShellProps = {
	target: CategorySettingsTarget;
	section: CategorySettingsSection;
	onClose: () => void;
	onForceClose: () => void;
};

export function CategorySettingsShell({
	target,
	section,
	onClose,
	onForceClose,
}: CategorySettingsShellProps) {
	const { t } = useTranslation();
	const { runIfAllowed } = useGroupSettingsUnsaved();

	const syncKey = `${target.groupId}:${target.categoryId}:${section}`;
	const [sectionOverride, setSectionOverride] = useState<{
		key: string;
		section: CategorySettingsSection;
	} | null>(null);
	const activeSection = sectionOverride?.key === syncKey ? sectionOverride.section : section;

	const handleSectionChange = (next: CategorySettingsSection) => {
		if (next === activeSection) {
			return;
		}
		runIfAllowed(() => setSectionOverride({ key: syncKey, section: next }));
	};

	return (
		<FullscreenSettingsShell
			titleId={CATEGORY_SETTINGS_TITLE_ID}
			srTitle={`${t("channel_category.settings.header_title")} — ${target.categoryName}`}
			onClose={onClose}
			sidebar={
				<CategorySettingsSidebar
					target={target}
					activeSection={activeSection}
					onSectionChange={handleSectionChange}
					onDeleted={onForceClose}
				/>
			}
		>
			{activeSection === "overview" ? (
				<CategorySettingsOverviewSection target={target} />
			) : (
				<CategorySettingsPermissionsSection target={target} />
			)}
		</FullscreenSettingsShell>
	);
}
