"use client";

import type { Group } from "@wichat/contracts";
import { useTranslation } from "react-i18next";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";

const LEAVE_GROUP_DIALOG_TITLE_ID = "wichat-leave-group-dialog-title";

type LeaveGroupConfirmDialogProps = {
	group: Group;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onConfirm: () => void;
	confirmDisabled?: boolean;
};

export function LeaveGroupConfirmDialog({
	group,
	open,
	onOpenChange,
	onConfirm,
	confirmDisabled,
}: LeaveGroupConfirmDialogProps) {
	const { t } = useTranslation();

	return (
		<ConfirmDialog
			open={open}
			onOpenChange={onOpenChange}
			titleId={LEAVE_GROUP_DIALOG_TITLE_ID}
			title={t("shell.leave_group_dialog_title", { name: group.name })}
			message={t("shell.leave_group_dialog_message", { name: group.name })}
			cancelLabel={t("shell.leave_group_cancel")}
			confirmLabel={t("shell.leave_group_action")}
			onConfirm={onConfirm}
			confirmDisabled={confirmDisabled}
			destructive
		/>
	);
}
