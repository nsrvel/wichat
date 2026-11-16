"use client";

import { useTranslation } from "react-i18next";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";

const DELETE_CATEGORY_DIALOG_TITLE_ID = "wichat-delete-category-dialog-title";

type DeleteCategoryConfirmDialogProps = {
	categoryName: string;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onConfirm: () => void;
	confirmDisabled?: boolean;
};

export function DeleteCategoryConfirmDialog({
	categoryName,
	open,
	onOpenChange,
	onConfirm,
	confirmDisabled,
}: DeleteCategoryConfirmDialogProps) {
	const { t } = useTranslation();

	return (
		<ConfirmDialog
			open={open}
			onOpenChange={onOpenChange}
			titleId={DELETE_CATEGORY_DIALOG_TITLE_ID}
			title={t("channel_category.settings.delete_confirm_title")}
			message={t("channel_category.settings.delete_confirm_message", { name: categoryName })}
			cancelLabel={t("channel_category.settings.delete_confirm_cancel")}
			confirmLabel={t("channel_category.settings.delete_confirm_action")}
			onConfirm={onConfirm}
			confirmDisabled={confirmDisabled}
			destructive
		/>
	);
}
