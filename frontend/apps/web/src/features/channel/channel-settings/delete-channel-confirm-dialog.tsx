"use client";

import { useTranslation } from "react-i18next";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";

const DELETE_CHANNEL_DIALOG_TITLE_ID = "wichat-delete-channel-dialog-title";

type DeleteChannelConfirmDialogProps = {
	channelName: string;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onConfirm: () => void;
	confirmDisabled?: boolean;
};

export function DeleteChannelConfirmDialog({
	channelName,
	open,
	onOpenChange,
	onConfirm,
	confirmDisabled,
}: DeleteChannelConfirmDialogProps) {
	const { t } = useTranslation();

	return (
		<ConfirmDialog
			open={open}
			onOpenChange={onOpenChange}
			titleId={DELETE_CHANNEL_DIALOG_TITLE_ID}
			title={t("channel.channel_settings.delete_confirm_title")}
			message={t("channel.channel_settings.delete_confirm_message", { name: channelName })}
			cancelLabel={t("channel.channel_settings.delete_confirm_cancel")}
			confirmLabel={t("channel.channel_settings.delete_confirm_action")}
			onConfirm={onConfirm}
			confirmDisabled={confirmDisabled}
			destructive
		/>
	);
}
