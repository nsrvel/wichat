"use client";

import type { Group } from "@wichat/contracts";
import {
	BellOff,
	CirclePlus,
	Fingerprint,
	FolderPlus,
	LogOut,
	Pencil,
	Settings,
	Shield,
	UserPlus,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { GroupRailMenuCheckboxItem } from "@/features/shell/components/group-rail-menu/group-rail-menu-checkbox-item";
import { GroupRailNotificationSettingsSubmenu } from "@/features/shell/components/group-rail-menu/group-rail-notification-settings-submenu";
import { useGroupMenuKit } from "@/features/shell/components/group-rail-menu/group-menu-kit";
import {
	menuDestructiveItemClass,
	menuItemClass,
	menuSeparatorClass,
} from "@/features/shell/components/group-rail-menu/menu-chrome";
import { useGroupSettings } from "@/features/group-settings/group-settings-provider";
import { useInviteDialog } from "@/features/invite/invite-dialog-provider";
import type { useGroupMenuController } from "@/features/shell/components/group-rail-menu/use-group-menu-controller";

const menuIconClass = "h-4 w-4 shrink-0 opacity-70";

type GroupHeaderDropdownContentProps = {
	group: Group;
	closeMenu: () => void;
	controller: ReturnType<typeof useGroupMenuController>;
};

export function GroupHeaderDropdownContent({
	group,
	closeMenu,
	controller,
}: GroupHeaderDropdownContentProps) {
	const { t } = useTranslation();
	const kit = useGroupMenuKit();
	const { openInvite } = useInviteDialog();
	const { openGroupSettings } = useGroupSettings();

	const openGroupSettingsForGroup = () => {
		openGroupSettings({
			groupId: group.id,
			groupName: group.name,
			logoUrl: group.logo_url,
			description: group.description,
			bannerUrl: group.banner_url,
			bannerColor: group.banner_color,
			isPrivate: group.is_private,
			role: group.role,
			section: "profile",
		});
		closeMenu();
	};
	const {
		manage,
		showMemberLeave,
		openLeaveDialog,
		hideMutedChannels,
		setHideMutedChannels,
		showComingSoon,
		copyGroupId,
	} = controller;

	return (
		<>
			<kit.Item
				className={menuItemClass}
				onClick={() => {
					openInvite({
						scope: "group",
						groupId: group.id,
						groupName: group.name,
					});
					closeMenu();
				}}
			>
				<UserPlus className={menuIconClass} aria-hidden />
				{t("shell.group_header_menu.invite")}
			</kit.Item>

			{manage ? (
				<>
					<kit.Separator className={menuSeparatorClass} />
					<kit.Item className={menuItemClass} onClick={openGroupSettingsForGroup}>
						<Settings className={menuIconClass} aria-hidden />
						{t("shell.group_header_menu.group_settings")}
					</kit.Item>
					<kit.Item className={menuItemClass} onClick={showComingSoon}>
						<CirclePlus className={menuIconClass} aria-hidden />
						{t("shell.group_header_menu.create_channel")}
					</kit.Item>
					<kit.Item className={menuItemClass} onClick={showComingSoon}>
						<FolderPlus className={menuIconClass} aria-hidden />
						{t("shell.group_header_menu.create_category")}
					</kit.Item>
				</>
			) : null}

			<kit.Separator className={menuSeparatorClass} />

			<GroupRailNotificationSettingsSubmenu />

			<GroupRailMenuCheckboxItem
				label={t("shell.group_rail_menu.hide_muted_channels")}
				checked={hideMutedChannels}
				onCheckedChange={setHideMutedChannels}
				leadingIcon={<BellOff className={menuIconClass} aria-hidden />}
			/>

			<kit.Separator className={menuSeparatorClass} />

			<kit.Item className={menuItemClass} onClick={showComingSoon}>
				<Shield className={menuIconClass} aria-hidden />
				{t("shell.group_rail_menu.privacy_settings")}
			</kit.Item>
			{manage ? (
				<kit.Item className={menuItemClass} onClick={openGroupSettingsForGroup}>
					<Pencil className={menuIconClass} aria-hidden />
					{t("shell.group_rail_menu.edit_group_profile")}
				</kit.Item>
			) : (
				<kit.Item className={menuItemClass} onClick={showComingSoon}>
					<Pencil className={menuIconClass} aria-hidden />
					{t("shell.group_rail_menu.edit_group_profile")}
				</kit.Item>
			)}

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
						<LogOut className={menuIconClass} aria-hidden />
						{t("shell.leave_group")}
					</kit.Item>
				</>
			) : null}

			<kit.Separator className={menuSeparatorClass} />

			<kit.Item className={menuItemClass} onClick={() => void copyGroupId(closeMenu)}>
				<Fingerprint className="h-3.5 w-3.5 shrink-0 opacity-45" aria-hidden />
				{t("shell.group_header_menu.copy_group_id")}
			</kit.Item>
		</>
	);
}
