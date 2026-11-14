"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useId } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
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
	authFormFooterButtonClass,
	authSubmitClass,
	authSubmitGhostClass,
} from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";

const PasswordConfirmSchema = z.object({
	currentPassword: z.string().min(1, "validation.password_required"),
});

type PasswordConfirmDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	title: string;
	subtitle?: string;
	confirmLabel: string;
	destructive?: boolean;
	isPending: boolean;
	error?: unknown;
	onConfirm: (currentPassword: string) => void;
};

export function PasswordConfirmDialog({
	open,
	onOpenChange,
	title,
	subtitle,
	confirmLabel,
	destructive = false,
	isPending,
	error,
	onConfirm,
}: PasswordConfirmDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();

	const form = useForm<z.infer<typeof PasswordConfirmSchema>>({
		resolver: zodResolver(PasswordConfirmSchema),
		defaultValues: { currentPassword: "" },
	});

	useEffect(() => {
		if (!open) {
			form.reset({ currentPassword: "" });
		}
	}, [open, form]);

	useEffect(() => {
		if (error) {
			applySettingsFormError(form, error, t, {
				passwordField: "currentPassword",
			});
		}
	}, [error, form, t]);

	const handleSubmit = form.handleSubmit(({ currentPassword }) => {
		onConfirm(currentPassword);
	});

	return (
		<WiChatDialog
			elevated
			open={open}
			onOpenChange={onOpenChange}
			titleId={titleId}
			className="max-w-md"
		>
			<WiChatDialogHeader titleId={titleId} title={title} subtitle={subtitle} />
			<Form {...form}>
				<form onSubmit={handleSubmit}>
					<WiChatDialogBody>
						<PasswordField
							control={form.control}
							name="currentPassword"
							label={t("user_settings.account.current_password")}
							autoComplete="current-password"
							showLabel={t("login.show_password")}
							hideLabel={t("login.hide_password")}
							required
						/>
					</WiChatDialogBody>
					<WiChatDialogFooter>
						<WiChatDialogFooterActions
							backLabel={t("user_settings.account.mfa_modal.cancel")}
							onBack={() => onOpenChange(false)}
							primaryLabel={confirmLabel}
							primaryType="submit"
							primaryDisabled={isPending}
							backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
							primaryClassName={cn(
								authSubmitClass,
								authFormFooterButtonClass,
								destructive &&
									"bg-destructive text-destructive-foreground hover:bg-destructive/90",
							)}
						/>
					</WiChatDialogFooter>
				</form>
			</Form>
		</WiChatDialog>
	);
}
