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
import { useUpdateMyPasswordMutation } from "@/services/user/user.mutation";

const MIN_PASSWORD_LEN = 12;
const MAX_PASSWORD_LEN = 128;

type ChangePasswordDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	hasPassword?: boolean;
};

type ChangePasswordFormValues = {
	currentPassword: string;
	newPassword: string;
	confirmPassword: string;
};

function ChangePasswordDialogBody({
	onClose,
	titleId,
	hasPassword,
}: {
	onClose: () => void;
	titleId: string;
	hasPassword: boolean;
}) {
	const { t } = useTranslation();
	const formId = useId();
	const updateMutation = useUpdateMyPasswordMutation();

	const schema = z
		.object({
			currentPassword: hasPassword ? z.string().min(1, "validation.password_required") : z.string(),
			newPassword: z
				.string()
				.min(MIN_PASSWORD_LEN, "validation.password")
				.max(MAX_PASSWORD_LEN, "validation.password"),
			confirmPassword: z.string(),
		})
		.refine((data) => data.confirmPassword === data.newPassword, {
			message: "validation.password_mismatch",
			path: ["confirmPassword"],
		});

	const form = useForm<ChangePasswordFormValues>({
		resolver: zodResolver(schema),
		defaultValues: {
			currentPassword: "",
			newPassword: "",
			confirmPassword: "",
		},
	});

	const onSubmit = async (values: ChangePasswordFormValues) => {
		if (updateMutation.isPending) {
			return;
		}
		try {
			await updateMutation.mutateAsync({
				current_password: hasPassword ? values.currentPassword : "",
				new_password: values.newPassword,
			});
			toast.success(
				hasPassword
					? t("user_settings.account.change_password_success")
					: t("user_settings.account.setup_password_success"),
			);
			onClose();
		} catch (err) {
			applySettingsFormError(form, err, t, {
				passwordField: hasPassword ? "currentPassword" : undefined,
				paramFieldMap: {
					current_password: "currentPassword",
					new_password: "newPassword",
				},
			});
		}
	};

	return (
		<>
			<WiChatDialogHeader
				title={
					hasPassword
						? t("user_settings.account.change_password_title")
						: t("user_settings.account.setup_password_title")
				}
				subtitle={
					hasPassword
						? t("user_settings.account.change_password_subtitle")
						: t("user_settings.account.setup_password_subtitle")
				}
				titleId={titleId}
			/>
			<WiChatDialogBody className="pb-5">
				<Form {...form}>
					<form id={formId} onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
						{hasPassword ? (
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
						) : null}
						<PasswordField
							control={form.control}
							name="newPassword"
							label={t("user_settings.account.new_password")}
							autoComplete="new-password"
							required
							showLabel={t("login.show_password")}
							hideLabel={t("login.hide_password")}
							labelClassName={authFieldLabelClass}
							inputClassName={authInputClass}
						/>
						<PasswordField
							control={form.control}
							name="confirmPassword"
							label={t("user_settings.account.confirm_new_password")}
							autoComplete="new-password"
							required
							showLabel={t("login.show_password")}
							hideLabel={t("login.hide_password")}
							labelClassName={authFieldLabelClass}
							inputClassName={authInputClass}
						/>
					</form>
				</Form>
			</WiChatDialogBody>
			<WiChatDialogFooter>
				<WiChatDialogFooterActions
					backLabel={t("user_settings.account.change_username_cancel")}
					onBack={onClose}
					primaryLabel={t("user_settings.account.change_username_done")}
					primaryType="submit"
					primaryForm={formId}
					primaryDisabled={updateMutation.isPending}
					backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
					primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
				/>
			</WiChatDialogFooter>
		</>
	);
}

export function ChangePasswordDialog({
	open,
	onOpenChange,
	hasPassword = true,
}: ChangePasswordDialogProps) {
	const titleId = useId();

	return (
		<WiChatDialog elevated open={open} onOpenChange={onOpenChange} titleId={titleId}>
			{open ? (
				<ChangePasswordDialogBody
					hasPassword={hasPassword}
					onClose={() => onOpenChange(false)}
					titleId={titleId}
				/>
			) : null}
		</WiChatDialog>
	);
}
