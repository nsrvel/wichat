"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useId } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";
import { PasswordField } from "@/components/form/password-field";
import { Form } from "@/components/ui/form";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
	WiChatDialogHeader,
} from "@/components/ui/wichat-dialog";
import { applySettingsFormError } from "@/features/user-settings/lib/settings-form-error";
import {
	authFieldLabelClass,
	authFormFooterButtonClass,
	authInputClass,
	authSubmitClass,
	authSubmitGhostClass,
} from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";
import {
	useRevokeMyDeviceMutation,
	useRevokeOtherDevicesMutation,
} from "@/services/user/user.mutation";

export type DeviceLogoutTarget =
	| {
			kind: "single";
			deviceId: string;
			deviceLabel: string;
	  }
	| {
			kind: "all";
	  };

type DeviceLogoutDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	target: DeviceLogoutTarget | null;
	hasPassword: boolean;
	onSingleVerified?: (password: string) => void;
};

type DeviceLogoutFormValues = {
	currentPassword: string;
};

function DeviceLogoutDialogBody({
	target,
	hasPassword,
	titleId,
	onClose,
	onSingleVerified,
}: {
	target: DeviceLogoutTarget;
	hasPassword: boolean;
	titleId: string;
	onClose: () => void;
	onSingleVerified?: (password: string) => void;
}) {
	const { t } = useTranslation();
	const formId = useId();
	const revokeDevice = useRevokeMyDeviceMutation();
	const revokeOthers = useRevokeOtherDevicesMutation();

	const schema = z.object({
		currentPassword: hasPassword ? z.string().min(1, "validation.password_required") : z.string(),
	});

	const form = useForm<DeviceLogoutFormValues>({
		resolver: zodResolver(schema),
		defaultValues: { currentPassword: "" },
	});

	const isSingle = target.kind === "single";
	const isPending = revokeDevice.isPending || revokeOthers.isPending;

	const subtitle = isSingle
		? hasPassword
			? t("user_settings.devices.logout_single_dialog_subtitle", { device: target.deviceLabel })
			: t("user_settings.devices.logout_single_dialog_subtitle_no_pw", {
					device: target.deviceLabel,
				})
		: hasPassword
			? t("user_settings.devices.logout_all_dialog_subtitle")
			: t("user_settings.devices.logout_all_dialog_subtitle_no_pw");

	const onSubmit = (values: DeviceLogoutFormValues) => {
		if (isPending) {
			return;
		}
		const password = hasPassword ? values.currentPassword : "";
		if (target.kind === "single") {
			revokeDevice.mutate(
				{ id: target.deviceId, current_password: password },
				{
					onSuccess: () => {
						if (hasPassword && password) {
							onSingleVerified?.(password);
						}
						toast.success(t("user_settings.devices.revoke_success"));
						onClose();
					},
					onError: (err) => {
						applySettingsFormError(form, err, t, { passwordField: "currentPassword" });
					},
				},
			);
			return;
		}
		revokeOthers.mutate(
			{ current_password: password },
			{
				onSuccess: () => {
					toast.success(t("user_settings.devices.revoke_all_success"));
					onClose();
				},
				onError: (err) => {
					applySettingsFormError(form, err, t, { passwordField: "currentPassword" });
				},
			},
		);
	};

	return (
		<>
			<WiChatDialogHeader
				title={
					isSingle
						? t("user_settings.devices.logout_single_dialog_title")
						: t("user_settings.devices.logout_all_dialog_title")
				}
				subtitle={subtitle}
				titleId={titleId}
			/>
			<Form {...form}>
				{hasPassword ? (
					<WiChatDialogBody className="pb-5">
						<form id={formId} onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
							<PasswordField
								control={form.control}
								name="currentPassword"
								label={t("user_settings.account.current_password")}
								autoComplete="current-password"
								required
								showLabel={t("login.show_password")}
								hideLabel={t("login.hide_password")}
								labelClassName={authFieldLabelClass}
								inputClassName={authInputClass}
							/>
						</form>
					</WiChatDialogBody>
				) : (
					<form id={formId} onSubmit={form.handleSubmit(onSubmit)} />
				)}
			</Form>
			<WiChatDialogFooter>
				<WiChatDialogFooterActions
					backLabel={t("user_settings.account.change_username_cancel")}
					onBack={onClose}
					primaryLabel={
						isSingle
							? t("user_settings.devices.logout_single_dialog_confirm")
							: t("user_settings.devices.logout_all_dialog_confirm")
					}
					primaryType="submit"
					primaryForm={formId}
					primaryDisabled={isPending}
					backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
					primaryClassName={cn(
						authSubmitClass,
						authFormFooterButtonClass,
						"bg-destructive text-destructive-foreground hover:bg-destructive/90",
					)}
				/>
			</WiChatDialogFooter>
		</>
	);
}

export function DeviceLogoutDialog({
	open,
	onOpenChange,
	target,
	hasPassword,
	onSingleVerified,
}: DeviceLogoutDialogProps) {
	const titleId = useId();

	return (
		<WiChatDialog
			elevated
			open={open && target !== null}
			onOpenChange={onOpenChange}
			titleId={titleId}
		>
			{open && target ? (
				<DeviceLogoutDialogBody
					key={target.kind === "single" ? target.deviceId : "all"}
					target={target}
					hasPassword={hasPassword}
					titleId={titleId}
					onClose={() => onOpenChange(false)}
					onSingleVerified={onSingleVerified}
				/>
			) : null}
		</WiChatDialog>
	);
}
