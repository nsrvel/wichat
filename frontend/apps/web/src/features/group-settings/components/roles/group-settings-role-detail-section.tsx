"use client";

import { ChevronLeft } from "lucide-react";
import { useTranslation } from "react-i18next";
import {
	groupSettingsSectionHeaderClass,
	groupSettingsSectionLeadClass,
	groupSettingsSectionTitleClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import type { GroupSettingsRoleRow } from "@/features/group-settings/lib/group-settings-roles-view";
import { Button } from "@/components/ui/button";

type GroupSettingsRoleDetailSectionProps = {
	role: GroupSettingsRoleRow;
	onBack: () => void;
};

function roleTitle(role: GroupSettingsRoleRow, t: (key: string) => string): string {
	if (role.isSystem) {
		return t(`invite.role_${role.name}`);
	}
	return role.name;
}

export function GroupSettingsRoleDetailSection({
	role,
	onBack,
}: GroupSettingsRoleDetailSectionProps) {
	const { t } = useTranslation();
	const title = roleTitle(role, t);

	return (
		<div className="flex flex-col gap-6">
			<div className="flex flex-col gap-3">
				<Button
					type="button"
					variant="ghost"
					className="h-9 w-fit gap-1.5 px-2 text-muted-foreground hover:text-foreground"
					onClick={onBack}
				>
					<ChevronLeft className="h-4 w-4" aria-hidden />
					{t("group_settings.roles.back_to_list")}
				</Button>
				<div className={groupSettingsSectionHeaderClass}>
					<h2 className={groupSettingsSectionTitleClass}>{title}</h2>
					<p className={groupSettingsSectionLeadClass}>{t("group_settings.roles.detail_lead")}</p>
				</div>
			</div>
			<div className="rounded-xl border border-dashed border-border/80 bg-field/30 px-4 py-8 text-center text-sm text-muted-foreground">
				{t("group_settings.roles.detail_placeholder")}
			</div>
		</div>
	);
}
