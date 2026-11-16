"use client";

import { ContextMenu } from "@base-ui/react/context-menu";
import { ChevronRight } from "lucide-react";
import { useState, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
	type NotificationLevel,
	SidebarNotificationMenuItems,
} from "@/components/layout/channel-sidebar/sidebar-notification-menu-items";
import {
	contextMenuKit,
	GroupMenuKitProvider,
} from "@/features/shell/components/group-rail-menu/group-menu-kit";
import { GroupRailMenuCheckboxItem } from "@/features/shell/components/group-rail-menu/group-rail-menu-checkbox-item";
import {
	menuChevronClass,
	menuDestructiveItemClass,
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
	menuSeparatorClass,
	submenuTriggerClass,
} from "@/lib/menu-chrome";
import { cn } from "@/lib/utils";

type SidebarCategoryContextMenuProps = {
	categoryName: string;
	open: boolean;
	onToggle: () => void;
	onCollapseAll?: () => void;
	hasUnread?: boolean;
	canManage?: boolean;
	isFirst?: boolean;
	isLast?: boolean;
	onMarkAsRead?: () => void;
	onMoveToTop?: () => void;
	onMoveToBottom?: () => void;
	onEditCategory?: () => void;
	onDeleteCategory?: () => void;
	children: ReactNode;
};

export function SidebarCategoryContextMenu({
	categoryName,
	open,
	onToggle,
	onCollapseAll,
	hasUnread = false,
	canManage = false,
	isFirst = false,
	isLast = false,
	onMarkAsRead,
	onMoveToTop,
	onMoveToBottom,
	onEditCategory,
	onDeleteCategory,
	children,
}: SidebarCategoryContextMenuProps) {
	const { t } = useTranslation();
	const [muted, setMuted] = useState(false);
	const [notifLevel, setNotifLevel] = useState<NotificationLevel>("default");

	return (
		<ContextMenu.Root>
			<ContextMenu.Trigger className="block">{children}</ContextMenu.Trigger>
			<ContextMenu.Portal>
				<ContextMenu.Positioner className={menuPositionerClass}>
					<ContextMenu.Popup className={menuPopupClass}>
						<GroupMenuKitProvider kit={contextMenuKit}>
							<ContextMenu.Item
								disabled={!hasUnread}
								className={cn(
									menuItemClass,
									!hasUnread && "pointer-events-none text-muted-foreground/60",
								)}
								onClick={() => {
									if (hasUnread) {
										onMarkAsRead?.();
									}
								}}
							>
								{t("channel_category.context_menu.mark_as_read")}
							</ContextMenu.Item>

							<ContextMenu.Separator className={menuSeparatorClass} />

							<GroupRailMenuCheckboxItem
								label={t("channel_category.context_menu.collapse_category")}
								checked={!open}
								onCheckedChange={() => onToggle()}
							/>

							<ContextMenu.Item
								className={menuItemClass}
								onClick={() => {
									onCollapseAll?.();
								}}
							>
								{t("channel_category.context_menu.collapse_all")}
							</ContextMenu.Item>

							<ContextMenu.Separator className={menuSeparatorClass} />

							<SidebarNotificationMenuItems
								muted={muted}
								onMutedChange={setMuted}
								muteLabel={t("channel_category.context_menu.mute_category")}
								unmuteLabel={t("channel_category.context_menu.unmute_category")}
								notifLevel={notifLevel}
								onNotifLevelChange={setNotifLevel}
								notificationSettingsLabel={t("channel_category.context_menu.notification_settings")}
								defaultLevelLabel={t("channel_category.context_menu.use_server_default")}
							/>

							{canManage && (onMoveToTop || onMoveToBottom) ? (
								<>
									<ContextMenu.Separator className={menuSeparatorClass} />

									<ContextMenu.SubmenuRoot>
										<ContextMenu.SubmenuTrigger className={submenuTriggerClass}>
											<span className="min-w-0 flex-1">
												{t("channel_category.context_menu.move_to")}
											</span>
											<ChevronRight className={menuChevronClass} aria-hidden />
										</ContextMenu.SubmenuTrigger>
										<ContextMenu.Portal>
											<ContextMenu.Positioner
												className={menuPositionerClass}
												alignOffset={-4}
												sideOffset={-4}
											>
												<ContextMenu.Popup className={menuPopupClass}>
													<ContextMenu.Item
														className={menuItemClass}
														disabled={isFirst}
														onClick={() => onMoveToTop?.()}
													>
														{t("shell.group_rail_menu.move_first")}
													</ContextMenu.Item>
													<ContextMenu.Item
														className={menuItemClass}
														disabled={isLast}
														onClick={() => onMoveToBottom?.()}
													>
														{t("shell.group_rail_menu.move_last")}
													</ContextMenu.Item>
												</ContextMenu.Popup>
											</ContextMenu.Positioner>
										</ContextMenu.Portal>
									</ContextMenu.SubmenuRoot>
								</>
							) : null}

							{canManage ? (
								<>
									<ContextMenu.Separator className={menuSeparatorClass} />
									<ContextMenu.Item
										className={menuItemClass}
										onClick={() => {
											if (onEditCategory) {
												onEditCategory();
											} else {
												toast.info(categoryName);
											}
										}}
									>
										{t("channel_category.context_menu.edit_category")}
									</ContextMenu.Item>
									<ContextMenu.Item
										className={menuDestructiveItemClass}
										onClick={() => {
											if (onDeleteCategory) {
												onDeleteCategory();
											} else {
												toast.info(categoryName);
											}
										}}
									>
										{t("channel_category.context_menu.delete_category")}
									</ContextMenu.Item>
								</>
							) : null}
						</GroupMenuKitProvider>
					</ContextMenu.Popup>
				</ContextMenu.Positioner>
			</ContextMenu.Portal>
		</ContextMenu.Root>
	);
}
