"use client";

import { ContextMenu } from "@base-ui/react/context-menu";
import { Menu } from "@base-ui/react/menu";
import { Copy, MoreHorizontal, Pencil, Pin, PinOff, Reply, Trash2 } from "lucide-react";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import {
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
	menuSeparatorClass,
} from "@/lib/menu-chrome";
import { cn } from "@/lib/utils";

export type MessageMenuActionProps = {
	isMine: boolean;
	isPinned: boolean;
	canPin: boolean;
	canDelete: boolean;
	onReply: () => void;
	onEdit?: () => void;
	onTogglePin?: () => void;
	onCopy?: () => void;
	onDelete?: () => void;
};

type MenuPrimitive = {
	Item: typeof Menu.Item | typeof ContextMenu.Item;
	Separator: typeof Menu.Separator | typeof ContextMenu.Separator;
};

function MessageMenuItems({
	primitive,
	actions,
}: {
	primitive: MenuPrimitive;
	actions: MessageMenuActionProps;
}) {
	const { t } = useTranslation();
	const { Item, Separator } = primitive;
	const { isMine, isPinned, canPin, canDelete, onReply, onEdit, onTogglePin, onCopy, onDelete } =
		actions;

	return (
		<>
			<Item className={menuItemClass} onClick={onReply}>
				<Reply className="size-4 shrink-0 text-muted-foreground" />
				<span>{t("chat.reply_action")}</span>
			</Item>

			{isMine && onEdit ? (
				<Item className={menuItemClass} onClick={onEdit}>
					<Pencil className="size-4 shrink-0 text-muted-foreground" />
					<span>{t("chat.edit_message")}</span>
				</Item>
			) : null}

			{canPin && onTogglePin ? (
				<Item className={menuItemClass} onClick={onTogglePin}>
					{isPinned ? (
						<PinOff className="size-4 shrink-0 text-muted-foreground" />
					) : (
						<Pin className="size-4 shrink-0 text-muted-foreground" />
					)}
					<span>{isPinned ? t("chat.unpin_message") : t("chat.pin_message")}</span>
				</Item>
			) : null}

			{onCopy ? (
				<Item className={menuItemClass} onClick={onCopy}>
					<Copy className="size-4 shrink-0 text-muted-foreground" />
					<span>{t("chat.copy_message")}</span>
				</Item>
			) : null}

			{canDelete && onDelete ? (
				<>
					<Separator className={menuSeparatorClass} />
					<Item
						className={cn(
							menuItemClass,
							"text-destructive data-highlighted:bg-destructive/15 data-highlighted:text-destructive",
						)}
						onClick={onDelete}
					>
						<Trash2 className="size-4 shrink-0" />
						<span>{t("chat.delete_message")}</span>
					</Item>
				</>
			) : null}
		</>
	);
}

const CONTEXT_PRIMITIVE: MenuPrimitive = {
	Item: ContextMenu.Item,
	Separator: ContextMenu.Separator,
};

const DROPDOWN_PRIMITIVE: MenuPrimitive = {
	Item: Menu.Item,
	Separator: Menu.Separator,
};

export function MessageBubbleContextMenu({
	children,
	...actions
}: MessageMenuActionProps & { children: ReactNode }) {
	return (
		<ContextMenu.Root>
			<ContextMenu.Trigger className="block">{children}</ContextMenu.Trigger>
			<ContextMenu.Portal>
				<ContextMenu.Positioner className={menuPositionerClass}>
					<ContextMenu.Popup className={menuPopupClass}>
						<MessageMenuItems primitive={CONTEXT_PRIMITIVE} actions={actions} />
					</ContextMenu.Popup>
				</ContextMenu.Positioner>
			</ContextMenu.Portal>
		</ContextMenu.Root>
	);
}

export function MessageMoreMenuButton(actions: MessageMenuActionProps) {
	const { t } = useTranslation();
	return (
		<Menu.Root>
			<Menu.Trigger
				type="button"
				className="flex size-7 items-center justify-center rounded-md text-muted-foreground transition hover:bg-sidebar-accent hover:text-sidebar-accent-foreground data-[popup-open]:bg-sidebar-accent data-[popup-open]:text-sidebar-accent-foreground"
				aria-label={t("chat.more")}
				title={t("chat.more")}
			>
				<MoreHorizontal className="size-4" />
			</Menu.Trigger>
			<Menu.Portal>
				<Menu.Positioner side="bottom" align="end" sideOffset={6} className={menuPositionerClass}>
					<Menu.Popup className={menuPopupClass}>
						<MessageMenuItems primitive={DROPDOWN_PRIMITIVE} actions={actions} />
					</Menu.Popup>
				</Menu.Positioner>
			</Menu.Portal>
		</Menu.Root>
	);
}
