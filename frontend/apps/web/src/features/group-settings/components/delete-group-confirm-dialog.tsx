"use client";

import { useTranslation } from "react-i18next";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";

const DELETE_GROUP_DIALOG_TITLE_ID = "wichat-delete-group-dialog-title";

type DeleteGroupConfirmDialogProps = {
	groupName: string;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onConfirm: () => void;
	confirmDisabled?: boolean;
};

export function DeleteGroupConfirmDialog({
	groupName,
	open,
	onOpenChange,
	onConfirm,
	confirmDisabled,
}: DeleteGroupConfirmDialogProps) {
	const { t } = useTranslation();

	return (
		<ConfirmDialog
			open={open}
			onOpenChange={onOpenChange}
			titleId={DELETE_GROUP_DIALOG_TITLE_ID}
			title={t("shell.delete_group")}
			message={t("shell.delete_group_confirm", { name: groupName })}
			cancelLabel={t("shell.delete_group_cancel")}
			confirmLabel={t("shell.delete_group_action")}
			onConfirm={onConfirm}
			confirmDisabled={confirmDisabled}
			destructive
		/>
	);
}
