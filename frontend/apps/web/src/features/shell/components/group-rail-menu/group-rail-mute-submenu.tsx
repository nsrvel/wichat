"use client";

import { ChevronRight } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useGroupMenuKit } from "@/features/shell/components/group-rail-menu/group-menu-kit";
import {
	menuChevronClass,
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
	submenuTriggerClass,
} from "@/lib/menu-chrome";

const MUTE_DURATION_MESSAGE_KEYS = [
	"shell.group_rail_menu.mute_for_15_minutes",
	"shell.group_rail_menu.mute_for_1_hour",
	"shell.group_rail_menu.mute_for_3_hours",
	"shell.group_rail_menu.mute_for_8_hours",
	"shell.group_rail_menu.mute_for_24_hours",
	"shell.group_rail_menu.mute_until_manual",
] as const;

type GroupRailMuteSubmenuProps = {
	onPickDuration: () => void;
};

export function GroupRailMuteSubmenu({ onPickDuration }: GroupRailMuteSubmenuProps) {
	const { t } = useTranslation();
	const kit = useGroupMenuKit();

	return (
		<kit.SubmenuRoot>
			<kit.SubmenuTrigger className={submenuTriggerClass}>
				<span className="min-w-0 flex-1">{t("shell.group_rail_menu.mute_group")}</span>
				<ChevronRight className={menuChevronClass} aria-hidden />
			</kit.SubmenuTrigger>
			<kit.Portal>
				<kit.Positioner className={menuPositionerClass} alignOffset={-4} sideOffset={-4}>
					<kit.Popup className={menuPopupClass}>
						{MUTE_DURATION_MESSAGE_KEYS.map((messageKey) => (
							<kit.Item key={messageKey} className={menuItemClass} onClick={onPickDuration}>
								{t(messageKey)}
							</kit.Item>
						))}
					</kit.Popup>
				</kit.Positioner>
			</kit.Portal>
		</kit.SubmenuRoot>
	);
}
