"use client";

import { ApiError } from "@wichat/api-client/api-error";
import type { GroupMember } from "@wichat/contracts/group/member";
import { useId, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { CustomSelect, type SelectOption } from "@/components/ui/custom-select";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogHeader,
} from "@/components/ui/wichat-dialog";
import { authInputClass } from "@/lib/form-field-styles";
import {
	useBanGroupMembersMutation,
	useKickGroupMemberMutation,
} from "@/services/group/group.mutation";

type MemberKickDialogProps = {
	groupId: string;
	member: GroupMember | null;
	onClose: () => void;
};

/**
 * Confirmation dialog for kicking a single group member.
 */
export function MemberKickDialog({ groupId, member, onClose }: MemberKickDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();
	const kickMutation = useKickGroupMemberMutation(groupId);
	const memberName = member?.display_name.trim() || t("invite.member_name_fallback");

	const handleConfirm = async () => {
		if (!member) {
			return;
		}
		try {
			await kickMutation.mutateAsync(member.user_id);
			toast.success(t("group_settings.members.kick_dialog.success_toast", { name: memberName }));
			onClose();
		} catch (error) {
			toast.error(error instanceof ApiError ? error.message : t("common.error_generic"));
		}
	};

	return (
		<ConfirmDialog
			open={member !== null}
			onOpenChange={(open) => {
				if (!open && !kickMutation.isPending) {
					onClose();
				}
			}}
			titleId={titleId}
			title={t("group_settings.members.kick_dialog.title", { name: memberName })}
			message={t("group_settings.members.kick_dialog.message", { name: memberName })}
			cancelLabel={t("group_settings.members.kick_dialog.cancel")}
			confirmLabel={t("group_settings.members.kick_dialog.confirm")}
			confirmDisabled={kickMutation.isPending}
			destructive
			onConfirm={() => void handleConfirm()}
		/>
	);
}

type MemberBanDialogProps = {
	groupId: string;
	members: GroupMember[];
	onClose: () => void;
	onBanned?: () => void;
};

/**
 * Moderation dialog for banning one or multiple group members with an optional
 * reason and message-history deletion window.
 */
export function MemberBanDialog({ groupId, members, onClose, onBanned }: MemberBanDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();
	const reasonId = useId();
	const banMutation = useBanGroupMembersMutation(groupId);
	const [reason, setReason] = useState("");
	const [deleteSeconds, setDeleteSeconds] = useState("0");

	const isOpen = members.length > 0;
	const isBulk = members.length > 1;
	const firstMemberName = members[0]?.display_name.trim() || t("invite.member_name_fallback");

	const deleteOptions = useMemo<SelectOption[]>(
		() => [
			{ value: "0", label: t("group_settings.members.ban_dialog.delete_none") },
			{ value: "3600", label: t("group_settings.members.ban_dialog.delete_1h") },
			{ value: "86400", label: t("group_settings.members.ban_dialog.delete_24h") },
			{ value: "604800", label: t("group_settings.members.ban_dialog.delete_7d") },
		],
		[t],
	);

	const resetAndClose = () => {
		setReason("");
		setDeleteSeconds("0");
		onClose();
	};

	const handleSubmit = async () => {
		if (members.length === 0) {
			return;
		}
		try {
			const res = await banMutation.mutateAsync({
				user_ids: members.map((m) => m.user_id),
				reason: reason.trim(),
				delete_message_seconds: Number(deleteSeconds) || 0,
			});
			toast.success(
				isBulk
					? t("group_settings.members.ban_dialog.success_bulk_toast", {
							count: res.banned_count || members.length,
						})
					: t("group_settings.members.ban_dialog.success_toast", { name: firstMemberName }),
			);
			onBanned?.();
			resetAndClose();
		} catch (error) {
			toast.error(error instanceof ApiError ? error.message : t("common.error_generic"));
		}
	};

	return (
		<WiChatDialog
			open={isOpen}
			elevated
			onOpenChange={(open) => {
				if (!open && !banMutation.isPending) {
					resetAndClose();
				}
			}}
			titleId={titleId}
		>
			<WiChatDialogHeader
				titleId={titleId}
				title={
					isBulk
						? t("group_settings.members.ban_dialog.title_bulk", { count: members.length })
						: t("group_settings.members.ban_dialog.title", { name: firstMemberName })
				}
				subtitle={
					isBulk
						? t("group_settings.members.ban_dialog.message_bulk", { count: members.length })
						: t("group_settings.members.ban_dialog.message", { name: firstMemberName })
				}
			/>
			<WiChatDialogBody className="space-y-4">
				<div className="space-y-1.5">
					<Label htmlFor={reasonId} className="text-xs font-medium text-foreground">
						{t("group_settings.members.ban_dialog.reason_label")}
					</Label>
					<Input
						id={reasonId}
						value={reason}
						onChange={(e) => setReason(e.target.value)}
						placeholder={t("group_settings.members.ban_dialog.reason_placeholder")}
						maxLength={512}
						className={authInputClass}
						autoComplete="off"
					/>
				</div>

				<div className="space-y-1.5">
					<Label className="text-xs font-medium text-foreground">
						{t("group_settings.members.ban_dialog.delete_history_label")}
					</Label>
					<CustomSelect
						value={deleteSeconds}
						placeholder={t("group_settings.members.ban_dialog.delete_none")}
						onChange={setDeleteSeconds}
						options={deleteOptions}
						triggerClassName={authInputClass}
						portaled
					/>
				</div>
			</WiChatDialogBody>
			<WiChatDialogFooter>
				<Button
					type="button"
					variant="ghost"
					disabled={banMutation.isPending}
					onClick={resetAndClose}
				>
					{t("group_settings.members.ban_dialog.cancel")}
				</Button>
				<Button
					type="button"
					variant="destructive"
					disabled={banMutation.isPending}
					onClick={() => void handleSubmit()}
				>
					{t("group_settings.members.ban_dialog.confirm")}
				</Button>
			</WiChatDialogFooter>
		</WiChatDialog>
	);
}
