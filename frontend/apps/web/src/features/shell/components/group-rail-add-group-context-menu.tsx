"use client";

import { ContextMenu } from "@base-ui/react/context-menu";
import { useRef, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import {
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
} from "@/features/shell/components/group-rail-menu/menu-chrome";
import { WICHAT_CONTEXT_MENU_ATTR } from "@/lib/context-menu-chrome";

type GroupRailAddGroupContextMenuProps = {
	onJoin: () => void;
	onCreate: () => void;
	children: ReactNode;
};

export function GroupRailAddGroupContextMenu({
	onJoin,
	onCreate,
	children,
}: GroupRailAddGroupContextMenuProps) {
	const { t } = useTranslation();
	const menuActionsRef = useRef<ContextMenu.Root.Actions>(null);

	const closeMenu = () => {
		menuActionsRef.current?.close();
	};

	return (
		<ContextMenu.Root actionsRef={menuActionsRef}>
			<ContextMenu.Trigger
				className="flex items-center justify-center"
				{...{ [WICHAT_CONTEXT_MENU_ATTR]: "" }}
			>
				{children}
			</ContextMenu.Trigger>
			<ContextMenu.Portal>
				<ContextMenu.Positioner className={menuPositionerClass}>
					<ContextMenu.Popup className={menuPopupClass}>
						<ContextMenu.Item
							className={menuItemClass}
							onClick={() => {
								onJoin();
								closeMenu();
							}}
						>
							{t("shell.add_group_menu_join")}
						</ContextMenu.Item>
						<ContextMenu.Item
							className={menuItemClass}
							onClick={() => {
								onCreate();
								closeMenu();
							}}
						>
							{t("shell.add_group_menu_create")}
						</ContextMenu.Item>
					</ContextMenu.Popup>
				</ContextMenu.Positioner>
			</ContextMenu.Portal>
		</ContextMenu.Root>
	);
}
