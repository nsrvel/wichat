"use client";

import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { groupSettingsFormFooterFloatAlignClass, groupSettingsFormFooterFloatColumnClass, groupSettingsFormFooterFloatShellClass, groupSettingsMainInnerClass } from "@/features/group-settings/lib/group-settings-chrome";
import { settingsFormFooterBarClass } from "@/lib/settings-chrome";
import { cn } from "@/lib/utils";

type GroupSettingsMembersSelectionFooterProps = {
	count: number;
	layout: "settings" | "workspace";
	onClear: () => void;
	onBan: () => void;
};

export function GroupSettingsMembersSelectionFooter({
	count,
	layout,
	onClear,
	onBan,
}: GroupSettingsMembersSelectionFooterProps) {
	const { t } = useTranslation();

	if (count <= 0) {
		return null;
	}

	const bar = (
		<div className={settingsFormFooterBarClass}>
			<p
				className="flex min-w-0 flex-1 items-center text-sm font-medium text-foreground"
				role="status"
			>
				{t("group_settings.members.selection_count", { count })}
			</p>
			<div className="flex shrink-0 items-center gap-2">
				<Button
					type="button"
					variant="ghost"
					className="text-muted-foreground hover:text-foreground"
					onClick={onClear}
				>
					{t("group_settings.members.clear_selection")}
				</Button>
				<Button type="button" variant="destructive" onClick={onBan}>
					{t("group_settings.members.ban_members")}
				</Button>
			</div>
		</div>
	);

	if (layout === "workspace") {
		return (
			<div className="sticky bottom-0 z-10 -mx-1 pt-3 pb-1">
				<div className="mx-auto w-full max-w-3xl shadow-elevated">{bar}</div>
			</div>
		);
	}

	return (
		<div className={groupSettingsFormFooterFloatShellClass}>
			<div className={groupSettingsFormFooterFloatAlignClass}>
				<div className="hidden lg:block" aria-hidden />
				<div className={groupSettingsFormFooterFloatColumnClass}>
					<div className={cn(groupSettingsMainInnerClass, "max-w-none lg:max-w-5xl")}>{bar}</div>
				</div>
			</div>
		</div>
	);
}
