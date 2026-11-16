"use client";

import { ApiError } from "@wichat/api-client/api-error";
import type { GroupMember } from "@wichat/contracts/group/member";
import { useId, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { CustomSelect, type SelectOption } from "@/components/ui/custom-select";
import { Label } from "@/components/ui/label";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
	WiChatDialogHeader,
} from "@/components/ui/wichat-dialog";
import {
	authFieldLabelClass,
	authInputClass,
	authSubmitClass,
	authSubmitGhostClass,
} from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";
import {
	useBanGroupMembersMutation,
	useKickGroupMemberMutation,
} from "@/services/group/group.mutation";

const BAN_DURATIONS = ["0", "3600", "86400", "604800", "2592000"] as const;
const moderationFooterButtonClass = "w-auto min-w-[7.5rem] max-w-[55%] shrink-0 px-8";
const destructiveSubmitClass = cn(
	authSubmitClass,
	"bg-destructive text-destructive-foreground hover:bg-destructive/90",
);
const descriptionInputClass = cn(authInputClass, "min-h-[6.5rem] resize-y py-2.5 sm:min-h-[7rem]");

type MemberKickDialogProps = {
	groupId: string;
	member: GroupMember | null;
	onClose: () => void;
};

/**
 * Moderation dialog for kicking a single group member with an optional
 * audit-log reason.
 */
export function MemberKickDialog({ groupId, member, onClose }: MemberKickDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();
	const reasonId = useId();
	const kickMutation = useKickGroupMemberMutation(groupId);
	const [reason, setReason] = useState("");
	const memberName = member?.display_name.trim() || t("invite.member_name_fallback");

	const resetAndClose = () => {
		setReason("");
		onClose();
	};

	const handleConfirm = async () => {
		if (!member) {
			return;
		}
		try {
			await kickMutation.mutateAsync({
				userId: member.user_id,
				reason: reason.trim(),
			});
			toast.success(t("group_settings.members.kick_dialog.success_toast", { name: memberName }));
			resetAndClose();
		} catch (error) {
			toast.error(error instanceof ApiError ? error.message : t("common.error_generic"));
		}
	};

	return (
		<WiChatDialog
			open={member !== null}
			onOpenChange={(open) => {
				if (!open && !kickMutation.isPending) {
					resetAndClose();
				}
			}}
			titleId={titleId}
			elevated
			className="max-w-md min-h-[min(360px,90dvh)]"
		>
			<WiChatDialogHeader
				title={t("group_settings.members.kick_dialog.title", { name: memberName })}
				subtitle={t("group_settings.members.kick_dialog.message", { name: memberName })}
				titleId={titleId}
				className="text-left"
			/>
			<WiChatDialogBody className="space-y-5">
				<div className="space-y-2">
					<Label htmlFor={reasonId} className={authFieldLabelClass}>
						{t("group_settings.members.kick_dialog.reason_label")}
					</Label>
					<textarea
						id={reasonId}
						value={reason}
						onChange={(e) => setReason(e.target.value)}
						maxLength={512}
						rows={4}
						placeholder={t("group_settings.members.kick_dialog.reason_placeholder")}
						className={descriptionInputClass}
					/>
				</div>
			</WiChatDialogBody>
			<WiChatDialogFooter className="pt-2">
				<WiChatDialogFooterActions
					backLabel={t("group_settings.members.kick_dialog.cancel")}
					onBack={resetAndClose}
					primaryLabel={t("group_settings.members.kick_dialog.confirm")}
					onPrimary={() => void handleConfirm()}
					primaryDisabled={kickMutation.isPending}
					backClassName={cn(authSubmitGhostClass, moderationFooterButtonClass)}
					primaryClassName={cn(destructiveSubmitClass, moderationFooterButtonClass)}
				/>
			</WiChatDialogFooter>
		</WiChatDialog>
	);
}

type MemberBanDialogProps = {
	groupId: string;
	members: GroupMember[];
	onClose: () => void;
	onBanned?: () => void;
};

/**
 * Moderation dialog for banning one or multiple group members with a duration
 * dropdown (including permanent) and an optional audit-log reason.
 */
export function MemberBanDialog({ groupId, members, onClose, onBanned }: MemberBanDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();
	const reasonId = useId();
	const banMutation = useBanGroupMembersMutation(groupId);
	const [durationSeconds, setDurationSeconds] = useState<string>("0");
	const [reason, setReason] = useState("");

	const isOpen = members.length > 0;
	const isBulk = members.length > 1;
	const firstMemberName = members[0]?.display_name.trim() || t("invite.member_name_fallback");

	const durationOptions = useMemo<SelectOption[]>(
		() =>
			BAN_DURATIONS.map((seconds) => ({
				value: seconds,
				label: t(`group_settings.members.ban_dialog.durations.${seconds}`),
			})),
		[t],
	);

	const resetAndClose = () => {
		setReason("");
		setDurationSeconds("0");
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
				delete_message_seconds: 0,
				duration_seconds: Number(durationSeconds) || 0,
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
			onOpenChange={(open) => {
				if (!open && !banMutation.isPending) {
					resetAndClose();
				}
			}}
			titleId={titleId}
			elevated
			className="max-w-md min-h-[min(420px,90dvh)]"
		>
			<WiChatDialogHeader
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
				titleId={titleId}
				className="text-left"
			/>
			<WiChatDialogBody className="space-y-5">
				<div className="space-y-2">
					<Label className={authFieldLabelClass}>
						{t("group_settings.members.ban_dialog.duration_label")}
					</Label>
					<CustomSelect
						value={durationSeconds}
						placeholder={t("group_settings.members.ban_dialog.durations.0")}
						onChange={setDurationSeconds}
						options={durationOptions}
						triggerClassName={authInputClass}
						portaled
					/>
				</div>

				<div className="space-y-2">
					<Label htmlFor={reasonId} className={authFieldLabelClass}>
						{t("group_settings.members.ban_dialog.reason_label")}
					</Label>
					<textarea
						id={reasonId}
						value={reason}
						onChange={(e) => setReason(e.target.value)}
						maxLength={512}
						rows={4}
						placeholder={t("group_settings.members.ban_dialog.reason_placeholder")}
						className={descriptionInputClass}
					/>
				</div>
			</WiChatDialogBody>
			<WiChatDialogFooter className="pt-2">
				<WiChatDialogFooterActions
					backLabel={t("group_settings.members.ban_dialog.cancel")}
					onBack={resetAndClose}
					primaryLabel={t("group_settings.members.ban_dialog.confirm")}
					onPrimary={() => void handleSubmit()}
					primaryDisabled={banMutation.isPending}
					backClassName={cn(authSubmitGhostClass, moderationFooterButtonClass)}
					primaryClassName={cn(destructiveSubmitClass, moderationFooterButtonClass)}
				/>
			</WiChatDialogFooter>
		</WiChatDialog>
	);
}
