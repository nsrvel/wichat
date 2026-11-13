"use client";

import type { Group } from "@wichat/contracts";
import { Fingerprint } from "lucide-react";
import { useTranslation } from "react-i18next";
import { GroupRailMenuCheckboxItem } from "@/features/shell/components/group-rail-menu/group-rail-menu-checkbox-item";
import { GroupRailMoveSubmenu } from "@/features/shell/components/group-rail-menu/group-rail-move-submenu";
import { GroupRailMuteSubmenu } from "@/features/shell/components/group-rail-menu/group-rail-mute-submenu";
import { GroupRailNotificationSettingsSubmenu } from "@/features/shell/components/group-rail-menu/group-rail-notification-settings-submenu";
import { useGroupMenuKit } from "@/features/shell/components/group-rail-menu/group-menu-kit";
import {
	menuDestructiveItemClass,
	menuItemClass,
	menuSeparatorClass,
} from "@/features/shell/components/group-rail-menu/menu-chrome";
import type { useGroupMenuController } from "@/features/shell/components/group-rail-menu/use-group-menu-controller";

type GroupMenuContentProps = {
	group: Group;
	orderedGroupIds: string[];
	closeMenu: () => void;
	controller: ReturnType<typeof useGroupMenuController>;
};

export function GroupMenuContent({
	group,
	orderedGroupIds,
	closeMenu,
	controller,
}: GroupMenuContentProps) {
	const { t } = useTranslation();
	const kit = useGroupMenuKit();
	const {
		showMemberLeave,
		openLeaveDialog,
		hideMutedChannels,
		setHideMutedChannels,
		showAllChannels,
		setShowAllChannels,
		showComingSoon,
		copyGroupId,
	} = controller;

	return (
		<>
			<kit.Item className={menuItemClass} onClick={showComingSoon}>
				{t("shell.group_rail_menu.mark_as_read")}
			</kit.Item>

			<kit.Separator className={menuSeparatorClass} />

			<GroupRailMuteSubmenu onPickDuration={showComingSoon} />

			<GroupRailNotificationSettingsSubmenu />

			<GroupRailMenuCheckboxItem
				label={t("shell.group_rail_menu.hide_muted_channels")}
				checked={hideMutedChannels}
				onCheckedChange={setHideMutedChannels}
			/>

			{showMemberLeave ? (
				<GroupRailMenuCheckboxItem
					label={t("shell.group_rail_menu.show_all_channels")}
					checked={showAllChannels}
					onCheckedChange={setShowAllChannels}
				/>
			) : null}

			<kit.Separator className={menuSeparatorClass} />

			<GroupRailMoveSubmenu
				groupId={group.id}
				orderedGroupIds={orderedGroupIds}
				onMoved={closeMenu}
			/>

			{showMemberLeave ? (
				<>
					<kit.Separator className={menuSeparatorClass} />
					<kit.Item
						className={menuDestructiveItemClass}
						onClick={() => {
							closeMenu();
							openLeaveDialog();
						}}
					>
						{t("shell.leave_group")}
					</kit.Item>
				</>
			) : null}

			<kit.Separator className={menuSeparatorClass} />

			<kit.Item className={menuItemClass} onClick={() => void copyGroupId(closeMenu)}>
				<Fingerprint className="size-3.5 shrink-0 opacity-45" aria-hidden />
				{t("shell.group_rail_menu.copy_group_id")}
			</kit.Item>
		</>
	);
}
