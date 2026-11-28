"use client";
import { WiChatContextMenuPositioner } from "@/components/ui/menu-positioner";

import { ContextMenu } from "@base-ui/react/context-menu";
import { ChevronRight } from "lucide-react";
import { useTranslation } from "react-i18next";
import { GroupRailMenuRadioItem } from "@/features/shell/components/group-rail-menu/group-rail-menu-radio-item";
import {
	menuChevronClass,
	menuHintClass,
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
	submenuTriggerClass,
} from "@/lib/menu-chrome";

export type NotificationLevel = "default" | "all" | "mentions" | "nothing";

export const NOTIFICATION_LEVEL_HINT_KEYS: Record<NotificationLevel, string> = {
	default: "shell.group_rail_menu.notification_level_mentions",
	all: "shell.group_rail_menu.notification_level_all",
	mentions: "shell.group_rail_menu.notification_level_mentions",
	nothing: "shell.group_rail_menu.notification_level_nothing",
};

export const MUTE_DURATION_KEYS = [
	"channel_category.context_menu.mute_15m",
	"channel_category.context_menu.mute_1h",
	"channel_category.context_menu.mute_3h",
	"channel_category.context_menu.mute_8h",
	"channel_category.context_menu.mute_24h",
	"channel_category.context_menu.mute_forever",
] as const;

type SidebarNotificationMenuItemsProps = {
	muted: boolean;
	onMutedChange: (muted: boolean) => void;
	muteLabel: string;
	unmuteLabel: string;
	notifLevel: NotificationLevel;
	onNotifLevelChange: (level: NotificationLevel) => void;
	notificationSettingsLabel: string;
	defaultLevelLabel: string;
};

export function SidebarNotificationMenuItems({
	muted,
	onMutedChange,
	muteLabel,
	unmuteLabel,
	notifLevel,
	onNotifLevelChange,
	notificationSettingsLabel,
	defaultLevelLabel,
}: SidebarNotificationMenuItemsProps) {
	const { t } = useTranslation();

	const levelOptions: Array<{ value: NotificationLevel; label: string; hint?: string }> = [
		{
			value: "default",
			label: defaultLevelLabel,
			hint: t("shell.group_rail_menu.notification_level_mentions"),
		},
		{ value: "all", label: t("shell.group_rail_menu.notification_level_all") },
		{ value: "mentions", label: t("shell.group_rail_menu.notification_level_mentions") },
		{ value: "nothing", label: t("shell.group_rail_menu.notification_level_nothing") },
	];

	return (
		<>
			<ContextMenu.SubmenuRoot>
				<ContextMenu.SubmenuTrigger className={submenuTriggerClass}>
					<span className="min-w-0 flex-1">{muted ? unmuteLabel : muteLabel}</span>
					<ChevronRight className={menuChevronClass} aria-hidden />
				</ContextMenu.SubmenuTrigger>
				<ContextMenu.Portal>
					<WiChatContextMenuPositioner
						className={menuPositionerClass}
						alignOffset={-4}
						sideOffset={-4}
					>
						<ContextMenu.Popup className={menuPopupClass}>
							{muted ? (
								<ContextMenu.Item className={menuItemClass} onClick={() => onMutedChange(false)}>
									{unmuteLabel}
								</ContextMenu.Item>
							) : (
								MUTE_DURATION_KEYS.map((key) => (
									<ContextMenu.Item
										key={key}
										className={menuItemClass}
										onClick={() => onMutedChange(true)}
									>
										{t(key)}
									</ContextMenu.Item>
								))
							)}
						</ContextMenu.Popup>
					</WiChatContextMenuPositioner>
				</ContextMenu.Portal>
			</ContextMenu.SubmenuRoot>

			<ContextMenu.SubmenuRoot>
				<ContextMenu.SubmenuTrigger className={submenuTriggerClass}>
					<div className="flex min-w-0 flex-1 flex-col items-start gap-0.5 py-0.5">
						<span className="leading-tight">{notificationSettingsLabel}</span>
						<span className={menuHintClass}>{t(NOTIFICATION_LEVEL_HINT_KEYS[notifLevel])}</span>
					</div>
					<ChevronRight className={menuChevronClass} aria-hidden />
				</ContextMenu.SubmenuTrigger>
				<ContextMenu.Portal>
					<WiChatContextMenuPositioner
						className={menuPositionerClass}
						alignOffset={-4}
						sideOffset={-4}
					>
						<ContextMenu.Popup className={menuPopupClass}>
							<ContextMenu.RadioGroup
								value={notifLevel}
								onValueChange={(next) => {
									if (
										next === "default" ||
										next === "all" ||
										next === "mentions" ||
										next === "nothing"
									) {
										onNotifLevelChange(next);
									}
								}}
							>
								{levelOptions.map((opt) => (
									<GroupRailMenuRadioItem
										key={opt.value}
										value={opt.value}
										label={opt.label}
										hint={opt.hint}
									/>
								))}
							</ContextMenu.RadioGroup>
						</ContextMenu.Popup>
					</WiChatContextMenuPositioner>
				</ContextMenu.Portal>
			</ContextMenu.SubmenuRoot>
		</>
	);
}
