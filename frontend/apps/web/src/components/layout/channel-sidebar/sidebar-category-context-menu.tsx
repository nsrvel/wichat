"use client";

import { ContextMenu } from "@base-ui/react/context-menu";
import { ChevronRight } from "lucide-react";
import { useState, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
	contextMenuKit,
	GroupMenuKitProvider,
} from "@/features/shell/components/group-rail-menu/group-menu-kit";
import { GroupRailMenuCheckboxItem } from "@/features/shell/components/group-rail-menu/group-rail-menu-checkbox-item";
import { GroupRailMenuRadioItem } from "@/features/shell/components/group-rail-menu/group-rail-menu-radio-item";
import {
	menuChevronClass,
	menuDestructiveItemClass,
	menuHintClass,
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
	menuSeparatorClass,
	submenuTriggerClass,
} from "@/lib/menu-chrome";
import { cn } from "@/lib/utils";

type NotificationLevel = "default" | "all" | "mentions" | "nothing";

const LEVEL_HINT_KEYS: Record<NotificationLevel, string> = {
	default: "shell.group_rail_menu.notification_level_mentions",
	all: "shell.group_rail_menu.notification_level_all",
	mentions: "shell.group_rail_menu.notification_level_mentions",
	nothing: "shell.group_rail_menu.notification_level_nothing",
};

const LEVEL_VALUES: Array<{ value: NotificationLevel; labelKey: string; hintKey?: string }> = [
	{
		value: "default",
		labelKey: "channel.category_context_menu.use_server_default",
		hintKey: "shell.group_rail_menu.notification_level_mentions",
	},
	{ value: "all", labelKey: "shell.group_rail_menu.notification_level_all" },
	{ value: "mentions", labelKey: "shell.group_rail_menu.notification_level_mentions" },
	{ value: "nothing", labelKey: "shell.group_rail_menu.notification_level_nothing" },
];

const MUTE_DURATION_KEYS = [
	"channel.category_context_menu.mute_15m",
	"channel.category_context_menu.mute_1h",
	"channel.category_context_menu.mute_3h",
	"channel.category_context_menu.mute_8h",
	"channel.category_context_menu.mute_24h",
	"channel.category_context_menu.mute_forever",
] as const;

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
								{t("channel.category_context_menu.mark_as_read")}
							</ContextMenu.Item>

							<ContextMenu.Separator className={menuSeparatorClass} />

							<GroupRailMenuCheckboxItem
								label={t("channel.category_context_menu.collapse_category")}
								checked={!open}
								onCheckedChange={() => onToggle()}
							/>

							<ContextMenu.Item
								className={menuItemClass}
								onClick={() => {
									onCollapseAll?.();
								}}
							>
								{t("channel.category_context_menu.collapse_all")}
							</ContextMenu.Item>

							<ContextMenu.Separator className={menuSeparatorClass} />

							<ContextMenu.SubmenuRoot>
								<ContextMenu.SubmenuTrigger className={submenuTriggerClass}>
									<span className="min-w-0 flex-1">
										{muted
											? t("channel.category_context_menu.unmute_category")
											: t("channel.category_context_menu.mute_category")}
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
											{muted ? (
												<ContextMenu.Item className={menuItemClass} onClick={() => setMuted(false)}>
													{t("channel.category_context_menu.unmute_category")}
												</ContextMenu.Item>
											) : (
												MUTE_DURATION_KEYS.map((key) => (
													<ContextMenu.Item
														key={key}
														className={menuItemClass}
														onClick={() => setMuted(true)}
													>
														{t(key)}
													</ContextMenu.Item>
												))
											)}
										</ContextMenu.Popup>
									</ContextMenu.Positioner>
								</ContextMenu.Portal>
							</ContextMenu.SubmenuRoot>

							<ContextMenu.SubmenuRoot>
								<ContextMenu.SubmenuTrigger className={submenuTriggerClass}>
									<div className="flex min-w-0 flex-1 flex-col items-start gap-0.5 py-0.5">
										<span className="leading-tight">
											{t("channel.category_context_menu.notification_settings")}
										</span>
										<span className={menuHintClass}>{t(LEVEL_HINT_KEYS[notifLevel])}</span>
									</div>
									<ChevronRight className={menuChevronClass} aria-hidden />
								</ContextMenu.SubmenuTrigger>
								<ContextMenu.Portal>
									<ContextMenu.Positioner
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
														setNotifLevel(next);
													}
												}}
											>
												{LEVEL_VALUES.map((opt) => (
													<GroupRailMenuRadioItem
														key={opt.value}
														value={opt.value}
														label={t(opt.labelKey)}
														hint={opt.hintKey ? t(opt.hintKey) : undefined}
													/>
												))}
											</ContextMenu.RadioGroup>
										</ContextMenu.Popup>
									</ContextMenu.Positioner>
								</ContextMenu.Portal>
							</ContextMenu.SubmenuRoot>

							{canManage && (onMoveToTop || onMoveToBottom) ? (
								<>
									<ContextMenu.Separator className={menuSeparatorClass} />

									<ContextMenu.SubmenuRoot>
										<ContextMenu.SubmenuTrigger className={submenuTriggerClass}>
											<span className="min-w-0 flex-1">
												{t("channel.category_context_menu.move_to")}
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
										{t("channel.category_context_menu.edit_category")}
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
										{t("channel.category_context_menu.delete_category")}
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
