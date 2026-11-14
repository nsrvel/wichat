"use client";

import { ChevronRight } from "lucide-react";
import { useState, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { GroupRailMenuCheckboxItem } from "@/features/shell/components/group-rail-menu/group-rail-menu-checkbox-item";
import { GroupRailMenuRadioItem } from "@/features/shell/components/group-rail-menu/group-rail-menu-radio-item";
import { useGroupMenuKit } from "@/features/shell/components/group-rail-menu/group-menu-kit";
import {
	menuChevronClass,
	menuHintClass,
	menuPopupClass,
	menuPositionerClass,
	menuSeparatorClass,
	submenuTriggerClass,
} from "@/lib/menu-chrome";

type NotificationLevel = "all" | "mentions" | "nothing";

const LEVEL_HINT_KEYS: Record<NotificationLevel, string> = {
	all: "shell.group_rail_menu.notification_level_all",
	mentions: "shell.group_rail_menu.notification_level_mentions",
	nothing: "shell.group_rail_menu.notification_level_nothing",
};

const LEVEL_VALUES: NotificationLevel[] = ["all", "mentions", "nothing"];

type GroupRailNotificationSettingsSubmenuProps = {
	leadingIcon?: ReactNode;
	/** Rail context menu shows hint; header dropdown is single-line. */
	showHint?: boolean;
	labelKey?: string;
};

// ponytail: persist when group notification prefs API exists — session-only toggles for now.
export function GroupRailNotificationSettingsSubmenu({
	leadingIcon,
	showHint = true,
	labelKey = "shell.group_rail_menu.notification_settings",
}: GroupRailNotificationSettingsSubmenuProps) {
	const { t } = useTranslation();
	const kit = useGroupMenuKit();
	const [level, setLevel] = useState<NotificationLevel>("mentions");
	const [suppressEveryone, setSuppressEveryone] = useState(false);
	const [suppressRoles, setSuppressRoles] = useState(false);
	const [suppressHighlights, setSuppressHighlights] = useState(false);
	const [muteNewEvents, setMuteNewEvents] = useState(false);
	const [mobilePush, setMobilePush] = useState(true);

	return (
		<kit.SubmenuRoot>
			<kit.SubmenuTrigger className={submenuTriggerClass}>
				{leadingIcon}
				{showHint ? (
					<div className="flex min-w-0 flex-1 flex-col items-start gap-1 py-0.5">
						<span className="leading-tight">{t(labelKey)}</span>
						<span className={menuHintClass}>{t(LEVEL_HINT_KEYS[level])}</span>
					</div>
				) : (
					<span className="min-w-0 flex-1">{t(labelKey)}</span>
				)}
				<ChevronRight className={menuChevronClass} aria-hidden />
			</kit.SubmenuTrigger>
			<kit.Portal>
				<kit.Positioner className={menuPositionerClass} alignOffset={-4} sideOffset={-4}>
					<kit.Popup className={menuPopupClass}>
						<kit.RadioGroup
							value={level}
							onValueChange={(next) => {
								if (next === "all" || next === "mentions" || next === "nothing") {
									setLevel(next);
								}
							}}
						>
							{LEVEL_VALUES.map((value) => (
								<GroupRailMenuRadioItem
									key={value}
									value={value}
									label={t(LEVEL_HINT_KEYS[value])}
								/>
							))}
						</kit.RadioGroup>

						<kit.Separator className={menuSeparatorClass} />

						<GroupRailMenuCheckboxItem
							label={t("shell.group_rail_menu.notification_suppress_everyone")}
							checked={suppressEveryone}
							onCheckedChange={setSuppressEveryone}
						/>
						<GroupRailMenuCheckboxItem
							label={t("shell.group_rail_menu.notification_suppress_roles")}
							checked={suppressRoles}
							onCheckedChange={setSuppressRoles}
						/>
						<GroupRailMenuCheckboxItem
							label={t("shell.group_rail_menu.notification_suppress_highlights")}
							checked={suppressHighlights}
							onCheckedChange={setSuppressHighlights}
						/>
						<GroupRailMenuCheckboxItem
							label={t("shell.group_rail_menu.notification_mute_new_events")}
							checked={muteNewEvents}
							onCheckedChange={setMuteNewEvents}
						/>

						<kit.Separator className={menuSeparatorClass} />

						<GroupRailMenuCheckboxItem
							label={t("shell.group_rail_menu.notification_mobile_push")}
							checked={mobilePush}
							onCheckedChange={setMobilePush}
						/>
					</kit.Popup>
				</kit.Positioner>
			</kit.Portal>
		</kit.SubmenuRoot>
	);
}
