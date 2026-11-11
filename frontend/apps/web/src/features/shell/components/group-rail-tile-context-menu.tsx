"use client";

import { ContextMenu } from "@base-ui/react/context-menu";
import type { Group } from "@wichat/contracts";
import { useRef, type ReactNode } from "react";
import { GroupMenuContent } from "@/features/shell/components/group-rail-menu/group-menu-content";
import {
	GroupMenuKitProvider,
	contextMenuKit,
} from "@/features/shell/components/group-rail-menu/group-menu-kit";
import {
	menuPopupClass,
	menuPositionerClass,
} from "@/features/shell/components/group-rail-menu/menu-chrome";
import { LeaveGroupConfirmDialog } from "@/features/shell/components/group-rail-menu/leave-group-confirm-dialog";
import { useGroupMenuController } from "@/features/shell/components/group-rail-menu/use-group-menu-controller";
import { WICHAT_CONTEXT_MENU_ATTR } from "@/lib/context-menu-chrome";

type GroupRailTileContextMenuProps = {
	group: Group;
	activeGroupId?: string;
	orderedGroupIds: string[];
	children: ReactNode;
};

export function GroupRailTileContextMenu({
	group,
	activeGroupId,
	orderedGroupIds,
	children,
}: GroupRailTileContextMenuProps) {
	const menuActionsRef = useRef<ContextMenu.Root.Actions>(null);
	const controller = useGroupMenuController({ group, activeGroupId });

	const closeMenu = () => {
		menuActionsRef.current?.close();
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
			<ContextMenu.Root
				actionsRef={menuActionsRef}
				onOpenChange={(open) => {
					if (!open) {
						controller.resetMenuState();
					}
				}}
			>
				<ContextMenu.Trigger
					className="flex items-center justify-center"
					{...{ [WICHAT_CONTEXT_MENU_ATTR]: "" }}
				>
					{children}
				</ContextMenu.Trigger>
				<ContextMenu.Portal>
					<ContextMenu.Positioner className={menuPositionerClass}>
						<ContextMenu.Popup className={menuPopupClass}>
							<GroupMenuKitProvider kit={contextMenuKit}>
								<GroupMenuContent
									group={group}
									orderedGroupIds={orderedGroupIds}
									closeMenu={closeMenu}
									controller={controller}
								/>
							</GroupMenuKitProvider>
						</ContextMenu.Popup>
					</ContextMenu.Positioner>
				</ContextMenu.Portal>
			</ContextMenu.Root>
		</>
	);
}
