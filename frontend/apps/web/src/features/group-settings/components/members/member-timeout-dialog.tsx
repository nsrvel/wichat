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
import { useTimeoutGroupMemberMutation } from "@/services/group/group.mutation";

const TIMEOUT_DURATIONS = [60, 300, 600, 3600, 86400, 604800] as const;
type TimeoutDuration = (typeof TIMEOUT_DURATIONS)[number];
const timeoutFooterButtonClass = "w-auto min-w-[7.5rem] max-w-[55%] shrink-0 px-8";
const descriptionInputClass = cn(authInputClass, "min-h-[6.5rem] resize-y py-2.5 sm:min-h-[7rem]");

type MemberTimeoutDialogProps = {
	groupId: string;
	member: GroupMember | null;
	onClose: () => void;
};

/**
 * Modal for timing out a group member with a duration dropdown
 * (60 secs .. 1 week) and an optional audit-log reason.
 */
export function MemberTimeoutDialog({ groupId, member, onClose }: MemberTimeoutDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();
	const reasonId = useId();
	const timeoutMutation = useTimeoutGroupMemberMutation(groupId);
	const [durationSeconds, setDurationSeconds] = useState<TimeoutDuration>(60);
	const [reason, setReason] = useState("");

	const durationOptions = useMemo<SelectOption[]>(
		() =>
			TIMEOUT_DURATIONS.map((seconds) => ({
				value: String(seconds),
				label: t(`group_settings.members.timeout_dialog.durations.${seconds}`),
			})),
		[t],
	);

	const memberName =
		member?.display_name.trim() || member?.username || t("invite.member_name_fallback");

	const resetAndClose = () => {
		setDurationSeconds(60);
		setReason("");
		onClose();
	};

	const handleSubmit = async () => {
		if (!member) {
			return;
		}
		try {
			await timeoutMutation.mutateAsync({
				userId: member.user_id,
				duration_seconds: durationSeconds,
				reason: reason.trim(),
			});
			toast.success(t("group_settings.members.timeout_dialog.success_toast", { name: memberName }));
			resetAndClose();
		} catch (error) {
			toast.error(error instanceof ApiError ? error.message : t("common.error_generic"));
		}
	};

	return (
		<WiChatDialog
			open={member !== null}
			onOpenChange={(open) => {
				if (!open && !timeoutMutation.isPending) {
					resetAndClose();
				}
			}}
			titleId={titleId}
			elevated
			className="max-w-md min-h-[min(420px,90dvh)]"
		>
			<WiChatDialogHeader
				title={t("group_settings.members.timeout_dialog.title", { name: memberName })}
				subtitle={t("group_settings.members.timeout_dialog.description")}
				titleId={titleId}
				className="text-left"
			/>
			<WiChatDialogBody className="space-y-5">
				<div className="space-y-2">
					<Label className={authFieldLabelClass}>
						{t("group_settings.members.timeout_dialog.duration_label")}
					</Label>
					<CustomSelect
						value={String(durationSeconds)}
						placeholder={t("group_settings.members.timeout_dialog.durations.60")}
						onChange={(val) => setDurationSeconds(Number(val) as TimeoutDuration)}
						options={durationOptions}
						triggerClassName={authInputClass}
						portaled
					/>
				</div>

				<div className="space-y-2">
					<Label htmlFor={reasonId} className={authFieldLabelClass}>
						{t("group_settings.members.timeout_dialog.reason_label")}
					</Label>
					<textarea
						id={reasonId}
						value={reason}
						onChange={(e) => setReason(e.target.value)}
						maxLength={512}
						rows={4}
						placeholder={t("group_settings.members.timeout_dialog.reason_placeholder")}
						className={descriptionInputClass}
					/>
				</div>
			</WiChatDialogBody>
			<WiChatDialogFooter className="pt-2">
				<WiChatDialogFooterActions
					backLabel={t("group_settings.members.timeout_dialog.cancel")}
					onBack={resetAndClose}
					primaryLabel={t("group_settings.members.timeout_dialog.confirm")}
					onPrimary={() => void handleSubmit()}
					primaryDisabled={timeoutMutation.isPending}
					backClassName={cn(authSubmitGhostClass, timeoutFooterButtonClass)}
					primaryClassName={cn(authSubmitClass, timeoutFooterButtonClass)}
				/>
			</WiChatDialogFooter>
		</WiChatDialog>
	);
}
