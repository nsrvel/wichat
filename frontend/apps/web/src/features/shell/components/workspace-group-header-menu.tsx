"use client";

import { ContextMenu } from "@base-ui/react/context-menu";
import { Menu } from "@base-ui/react/menu";
import type { Group } from "@wichat/contracts";
import { ChevronDown, ChevronUp, MessageSquarePlus, Search } from "lucide-react";
import { useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
	WorkspaceHeaderIconButton,
	workspaceHeaderIconClass,
} from "@/components/layout/workspace-header-icon-button";
import { GroupHeaderDropdownContent } from "@/features/shell/components/group-rail-menu/group-header-dropdown-content";
import { GroupMenuContent } from "@/features/shell/components/group-rail-menu/group-menu-content";
import {
	GroupMenuKitProvider,
	contextMenuKit,
	dropdownMenuKit,
} from "@/features/shell/components/group-rail-menu/group-menu-kit";
import {
	menuDropdownTriggerClass,
	menuDropdownTriggerChevronClass,
	menuDropdownTriggerOpenClass,
	menuPopupClass,
	menuPositionerClass,
} from "@/lib/menu-chrome";
import { LeaveGroupConfirmDialog } from "@/features/shell/components/group-rail-menu/leave-group-confirm-dialog";
import { useGroupMenuController } from "@/features/shell/components/group-rail-menu/use-group-menu-controller";
import { WICHAT_CONTEXT_MENU_ATTR } from "@/lib/context-menu-chrome";
import { cn } from "@/lib/utils";

type WorkspaceGroupHeaderMenuProps = {
	group: Group;
	activeGroupId?: string;
	orderedGroupIds: string[];
	className?: string;
	onSearchClick?: () => void;
	onNewMessageClick?: () => void;
};

export function WorkspaceGroupHeaderMenu({
	group,
	activeGroupId,
	orderedGroupIds,
	className,
	onSearchClick,
	onNewMessageClick,
}: WorkspaceGroupHeaderMenuProps) {
	const { t } = useTranslation();
	const [dropdownOpen, setDropdownOpen] = useState(false);
	const contextActionsRef = useRef<ContextMenu.Root.Actions>(null);
	const dropdownHandle = useMemo(() => Menu.createHandle(), []);
	const controller = useGroupMenuController({ group, activeGroupId });

	const closeContextMenu = () => {
		contextActionsRef.current?.close();
	};

	const closeDropdown = () => {
		setDropdownOpen(false);
	};

	const onMenuClosed = () => {
		controller.resetMenuState();
	};

	return (
		<>
			<LeaveGroupConfirmDialog
				group={group}
				open={controller.leaveDialogOpen}
				onOpenChange={controller.setLeaveDialogOpen}
				onConfirm={controller.confirmLeave}
				confirmDisabled={controller.leavePending}
			/>
			<div className={cn("flex w-full min-w-0 items-center gap-2", className)}>
				{/* Sibling roots: ContextMenu.Root embeds its own Menu — do not nest a second Menu.Root inside it. */}
				<Menu.Root
					handle={dropdownHandle}
					open={dropdownOpen}
					onOpenChange={(open) => {
						setDropdownOpen(open);
						if (open) {
							closeContextMenu();
						}
						if (!open) {
							onMenuClosed();
						}
					}}
				>
					<Menu.Portal>
						<Menu.Positioner className={menuPositionerClass} sideOffset={4} align="start">
							<Menu.Popup className={menuPopupClass}>
								<GroupMenuKitProvider kit={dropdownMenuKit}>
									<GroupHeaderDropdownContent
										group={group}
										closeMenu={closeDropdown}
										controller={controller}
									/>
								</GroupMenuKitProvider>
							</Menu.Popup>
						</Menu.Positioner>
					</Menu.Portal>
				</Menu.Root>

				<ContextMenu.Root
					actionsRef={contextActionsRef}
					onOpenChange={(open) => {
						if (open) {
							setDropdownOpen(false);
						}
						if (!open) {
							onMenuClosed();
						}
					}}
				>
					<div className="flex min-w-0 flex-1 items-center">
						<ContextMenu.Trigger
							className="flex max-w-full min-w-0"
							{...{ [WICHAT_CONTEXT_MENU_ATTR]: "" }}
						>
							<Menu.Trigger
								handle={dropdownHandle}
								nativeButton
								render={
									<button
										type="button"
										className={cn(
											menuDropdownTriggerClass,
											dropdownOpen && menuDropdownTriggerOpenClass,
										)}
										aria-label={t("shell.group_header_menu.label", { name: group.name })}
										aria-haspopup="menu"
										aria-expanded={dropdownOpen}
									>
										<span className="truncate text-[16px] font-bold tracking-tight text-sidebar-foreground lg:text-[15px]">
											{group.name}
										</span>
										{dropdownOpen ? (
											<ChevronUp
												className={menuDropdownTriggerChevronClass}
												strokeWidth={2.4}
												aria-hidden
											/>
										) : (
											<ChevronDown
												className={menuDropdownTriggerChevronClass}
												strokeWidth={2.4}
												aria-hidden
											/>
										)}
									</button>
								}
							/>
						</ContextMenu.Trigger>
					</div>
					<ContextMenu.Portal>
						<ContextMenu.Positioner className={menuPositionerClass}>
							<ContextMenu.Popup className={menuPopupClass}>
								<GroupMenuKitProvider kit={contextMenuKit}>
									<GroupMenuContent
										group={group}
										orderedGroupIds={orderedGroupIds}
										closeMenu={closeContextMenu}
										controller={controller}
									/>
								</GroupMenuKitProvider>
							</ContextMenu.Popup>
						</ContextMenu.Positioner>
					</ContextMenu.Portal>
				</ContextMenu.Root>

				<div className="ml-auto flex shrink-0 items-center gap-0.5">
					<WorkspaceHeaderIconButton aria-label="Search" onClick={onSearchClick}>
						<Search className={workspaceHeaderIconClass} strokeWidth={2} />
					</WorkspaceHeaderIconButton>
					<WorkspaceHeaderIconButton aria-label="New message" onClick={onNewMessageClick}>
						<MessageSquarePlus className={workspaceHeaderIconClass} strokeWidth={2} />
					</WorkspaceHeaderIconButton>
				</div>
			</div>
		</>
	);
}
