"use client";
import { WiChatMenuPositioner } from "@/components/ui/menu-positioner";

import { Menu } from "@base-ui/react/menu";
import { useId, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import {
	roleDisplayTitle,
	type GroupSettingsRoleRow,
} from "@/features/group-settings/lib/group-settings-roles-view";
import { ApiError } from "@wichat/api-client/api-error";
import {
	menuDestructiveItemClass,
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
	menuSeparatorClass,
} from "@/lib/menu-chrome";
import { cn } from "@/lib/utils";
import {
	useDeleteGroupRoleMutation,
	useDuplicateGroupRoleMutation,
} from "@/services/group-role/group-role.mutation";

export type RoleActionMenuState = {
	role: GroupSettingsRoleRow;
	anchor: HTMLElement | { getBoundingClientRect: () => DOMRect };
	align: "start" | "end";
};

export function createRoleContextMenuState(
	role: GroupSettingsRoleRow,
	clientX: number,
	clientY: number,
): RoleActionMenuState {
	return {
		role,
		anchor: {
			getBoundingClientRect: () => new DOMRect(clientX, clientY, 0, 0),
		},
		align: "start",
	};
}

type RoleActionMenuProps = {
	groupId: string;
	menuState: RoleActionMenuState | null;
	onCloseMenu: () => void;
	canManage: boolean;
	onDuplicated?: (roleId: string) => void;
	onDeleted?: (roleId: string) => void;
};

export function RoleActionMenu({
	groupId,
	menuState,
	onCloseMenu,
	canManage,
	onDuplicated,
	onDeleted,
}: RoleActionMenuProps) {
	const { t } = useTranslation();
	const deleteTitleId = useId();
	const [deleteTarget, setDeleteTarget] = useState<GroupSettingsRoleRow | null>(null);
	const duplicateMutation = useDuplicateGroupRoleMutation(groupId);
	const deleteMutation = useDeleteGroupRoleMutation(groupId);

	const handleDuplicate = async (role: GroupSettingsRoleRow) => {
		onCloseMenu();
		try {
			const created = await duplicateMutation.mutateAsync(role.id);
			toast.success(t("group_settings.roles.detail.duplicated_toast", { name: created.name }));
			onDuplicated?.(created.id);
		} catch (error) {
			toast.error(error instanceof ApiError ? error.message : t("common.error_generic"));
		}
	};

	const handleConfirmDelete = async () => {
		if (!deleteTarget) {
			return;
		}
		const target = deleteTarget;
		const label = roleDisplayTitle(target, t);
		try {
			await deleteMutation.mutateAsync(target.id);
			setDeleteTarget(null);
			toast.success(t("group_settings.roles.delete_dialog.deleted_toast", { name: label }));
			onDeleted?.(target.id);
		} catch (error) {
			toast.error(error instanceof ApiError ? error.message : t("common.error_generic"));
		}
	};

	return (
		<>
			<Menu.Root
				open={menuState !== null}
				onOpenChange={(open) => {
					if (!open) {
						onCloseMenu();
					}
				}}
			>
				<Menu.Portal>
					<WiChatMenuPositioner
						anchor={menuState?.anchor ?? null}
						className={menuPositionerClass}
						side="bottom"
						align={menuState?.align ?? "end"}
						sideOffset={4}
					>
						<Menu.Popup className={menuPopupClass}>
							{menuState ? (
								<>
									<Menu.Item
										className={cn(
											menuItemClass,
											"disabled:pointer-events-none disabled:opacity-45",
										)}
										disabled={!canManage || duplicateMutation.isPending}
										onClick={() => void handleDuplicate(menuState.role)}
									>
										{t("group_settings.roles.action_duplicate")}
									</Menu.Item>

									<Menu.Separator className={menuSeparatorClass} />

									<Menu.Item
										className={cn(
											menuDestructiveItemClass,
											"disabled:pointer-events-none disabled:opacity-45",
										)}
										disabled={!canManage || menuState.role.isSystem || menuState.role.isDefault}
										onClick={() => {
											const target = menuState.role;
											onCloseMenu();
											setDeleteTarget(target);
										}}
									>
										{t("group_settings.roles.action_delete")}
									</Menu.Item>
								</>
							) : null}
						</Menu.Popup>
					</WiChatMenuPositioner>
				</Menu.Portal>
			</Menu.Root>

			<ConfirmDialog
				open={deleteTarget !== null}
				onOpenChange={(open) => {
					if (!open) {
						setDeleteTarget(null);
					}
				}}
				titleId={deleteTitleId}
				title={t("group_settings.roles.delete_dialog.title")}
				message={t("group_settings.roles.delete_dialog.message", {
					name: deleteTarget ? roleDisplayTitle(deleteTarget, t) : "",
				})}
				cancelLabel={t("group_settings.roles.delete_dialog.cancel")}
				confirmLabel={t("group_settings.roles.delete_dialog.confirm")}
				confirmDisabled={deleteMutation.isPending}
				destructive
				onConfirm={() => void handleConfirmDelete()}
			/>
		</>
	);
}
