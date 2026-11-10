"use client";

import { ContextMenu } from "@base-ui/react/context-menu";
import { Menu } from "@base-ui/react/menu";
import { createContext, useContext, type ReactNode } from "react";

export type GroupMenuKit = {
	Item: typeof ContextMenu.Item;
	Separator: typeof ContextMenu.Separator;
	Group: typeof ContextMenu.Group;
	GroupLabel: typeof ContextMenu.GroupLabel;
	CheckboxItem: typeof ContextMenu.CheckboxItem;
	CheckboxItemIndicator: typeof ContextMenu.CheckboxItemIndicator;
	SubmenuRoot: typeof ContextMenu.SubmenuRoot;
	SubmenuTrigger: typeof ContextMenu.SubmenuTrigger;
	Portal: typeof ContextMenu.Portal;
	Positioner: typeof ContextMenu.Positioner;
	Popup: typeof ContextMenu.Popup;
	RadioGroup: typeof ContextMenu.RadioGroup;
	RadioItem: typeof ContextMenu.RadioItem;
	RadioItemIndicator: typeof ContextMenu.RadioItemIndicator;
};

export const contextMenuKit: GroupMenuKit = ContextMenu;
export const dropdownMenuKit = Menu as unknown as GroupMenuKit;

const GroupMenuKitContext = createContext<GroupMenuKit>(contextMenuKit);

export function GroupMenuKitProvider({
	kit,
	children,
}: {
	kit: GroupMenuKit;
	children: ReactNode;
}) {
	return <GroupMenuKitContext.Provider value={kit}>{children}</GroupMenuKitContext.Provider>;
}

export function useGroupMenuKit(): GroupMenuKit {
	return useContext(GroupMenuKitContext);
}
