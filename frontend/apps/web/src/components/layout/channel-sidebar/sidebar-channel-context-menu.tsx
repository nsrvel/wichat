"use client";

import { ContextMenu } from "@base-ui/react/context-menu";
import type { Channel } from "@wichat/contracts/chat/channel";
import type { ChannelCategory } from "@wichat/contracts/chat/channel-category";
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

type SidebarChannelContextMenuProps = {
	channel: Channel;
	categories: readonly ChannelCategory[];
	hasUnread?: boolean;
	canManage?: boolean;
	isFirstInCategory?: boolean;
	isLastInCategory?: boolean;
	onMarkAsRead?: () => void;
	onInvite?: () => void;
	onMoveToCategory?: (targetCategoryId: string | null) => void;
	onMoveToFirstInCategory?: () => void;
	onMoveToLastInCategory?: () => void;
	onEditChannel?: () => void;
	onDuplicateChannel?: () => void;
	onCreateChannel?: () => void;
	onDeleteChannel?: () => void;
	children: ReactNode;
};

export function SidebarChannelContextMenu({
	channel,
	categories,
	hasUnread = false,
	canManage = false,
	isFirstInCategory = false,
	isLastInCategory = false,
	onMarkAsRead,
	onInvite,
	onMoveToCategory,
	onMoveToFirstInCategory,
	onMoveToLastInCategory,
	onEditChannel,
	onDuplicateChannel,
	onCreateChannel,
	onDeleteChannel,
	children,
}: SidebarChannelContextMenuProps) {
	const { t } = useTranslation();
	const [muted, setMuted] = useState(false);
	const [notifLevel, setNotifLevel] = useState<NotificationLevel>("default");

	const currentCategoryId = channel.category_id || null;
	const canMoveToCategory = Boolean(onMoveToCategory) && categories.length > 0;
	const canMove =
		canManage &&
		(canMoveToCategory || Boolean(onMoveToFirstInCategory) || Boolean(onMoveToLastInCategory));

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
								{t("channel.channel_context_menu.mark_as_read")}
							</ContextMenu.Item>

							<ContextMenu.Separator className={menuSeparatorClass} />

							{onInvite ? (
								<>
									<ContextMenu.Item className={menuItemClass} onClick={onInvite}>
										{t("channel.channel_context_menu.invite_to_channel")}
									</ContextMenu.Item>
									<ContextMenu.Separator className={menuSeparatorClass} />
								</>
							) : null}

							<SidebarNotificationMenuItems
								muted={muted}
								onMutedChange={setMuted}
								muteLabel={t("channel.channel_context_menu.mute_channel")}
								unmuteLabel={t("channel.channel_context_menu.unmute_channel")}
								notifLevel={notifLevel}
								onNotifLevelChange={setNotifLevel}
								notificationSettingsLabel={t("channel.channel_context_menu.notification_settings")}
								defaultLevelLabel={t("channel.channel_context_menu.use_category_default")}
							/>

							{canMove ? (
								<>
									<ContextMenu.Separator className={menuSeparatorClass} />

									<ContextMenu.SubmenuRoot>
										<ContextMenu.SubmenuTrigger className={submenuTriggerClass}>
											<span className="min-w-0 flex-1">
												{t("channel.channel_context_menu.move_to")}
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
													{canMoveToCategory && onMoveToCategory ? (
														<>
															<ContextMenu.SubmenuRoot>
																<ContextMenu.SubmenuTrigger className={submenuTriggerClass}>
																	<span className="min-w-0 flex-1">
																		{t("channel.channel_context_menu.category")}
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
																			{categories.map((cat) => {
																				const isCurrent = cat.id === currentCategoryId;
																				return (
																					<ContextMenu.Item
																						key={cat.id}
																						disabled={isCurrent}
																						className={cn(
																							menuItemClass,
																							isCurrent &&
																								"pointer-events-none text-muted-foreground/60",
																						)}
																						onClick={() => onMoveToCategory(cat.id)}
																					>
																						<span className="truncate">{cat.name}</span>
																					</ContextMenu.Item>
																				);
																			})}
																		</ContextMenu.Popup>
																	</ContextMenu.Positioner>
																</ContextMenu.Portal>
															</ContextMenu.SubmenuRoot>
															<ContextMenu.Separator className={menuSeparatorClass} />
														</>
													) : null}

													<ContextMenu.Item
														disabled={isFirstInCategory}
														className={cn(
															menuItemClass,
															isFirstInCategory && "pointer-events-none text-muted-foreground/60",
														)}
														onClick={() => onMoveToFirstInCategory?.()}
													>
														{t("channel.channel_context_menu.first_in_category")}
													</ContextMenu.Item>
													<ContextMenu.Item
														disabled={isLastInCategory}
														className={cn(
															menuItemClass,
															isLastInCategory && "pointer-events-none text-muted-foreground/60",
														)}
														onClick={() => onMoveToLastInCategory?.()}
													>
														{t("channel.channel_context_menu.last_in_category")}
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
											if (onEditChannel) {
												onEditChannel();
											} else {
												toast.info(t("channel.channel_context_menu.coming_soon"));
											}
										}}
									>
										{t("channel.channel_context_menu.edit_channel")}
									</ContextMenu.Item>
									<ContextMenu.Item
										className={menuItemClass}
										onClick={() => {
											if (onDuplicateChannel) {
												onDuplicateChannel();
											} else {
												toast.info(t("channel.channel_context_menu.coming_soon"));
											}
										}}
									>
										{t("channel.channel_context_menu.duplicate_channel")}
									</ContextMenu.Item>
									<ContextMenu.Item
										className={menuItemClass}
										onClick={() => {
											if (onCreateChannel) {
												onCreateChannel();
											} else {
												toast.info(t("channel.channel_context_menu.coming_soon"));
											}
										}}
									>
										{t("channel.channel_context_menu.create_channel")}
									</ContextMenu.Item>
									<ContextMenu.Item
										className={menuDestructiveItemClass}
										onClick={() => {
											if (onDeleteChannel) {
												onDeleteChannel();
											} else {
												toast.info(t("channel.channel_context_menu.coming_soon"));
											}
										}}
									>
										{t("channel.channel_context_menu.delete_channel")}
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
