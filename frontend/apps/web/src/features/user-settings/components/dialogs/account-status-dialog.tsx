"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import { AlertTriangle } from "lucide-react";
import { useRouter } from "next/navigation";
import { useId } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";
import { PasswordField } from "@/components/form/password-field";
import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
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
import { myGroupsQueryOpts } from "@/services/group/group.query";
import {
	useDeleteMyAccountMutation,
	useDisableMyAccountMutation,
} from "@/services/user/user.mutation";

export type AccountStatusDialogMode = "disable" | "delete";

type AccountStatusDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	mode: AccountStatusDialogMode;
	hasPassword: boolean;
	onAccountClosed?: () => void;
};

type AccountStatusFormValues = {
	currentPassword: string;
	confirmText: string;
};

function AccountStatusDialogBody({
	mode,
	hasPassword,
	titleId,
	onClose,
	onAccountClosed,
}: {
	mode: AccountStatusDialogMode;
	hasPassword: boolean;
	titleId: string;
	onClose: () => void;
	onAccountClosed?: () => void;
}) {
	const { t } = useTranslation();
	const router = useRouter();
	const formId = useId();

	const { data: groups = [] } = useQuery(myGroupsQueryOpts());
	const ownedGroups = groups.filter((g) => g.role === "owner");

	const disableMutation = useDisableMyAccountMutation();
	const deleteMutation = useDeleteMyAccountMutation();
	const pending = disableMutation.isPending || deleteMutation.isPending;

	const isDelete = mode === "delete";
	const confirmWord = t("user_settings.account.delete_confirm_word");

	const schema = z.object({
		currentPassword: hasPassword ? z.string().min(1, "validation.password_required") : z.string(),
		confirmText: isDelete
			? z
					.string()
					.refine(
						(val) => val.trim().toUpperCase() === confirmWord.toUpperCase(),
						"validation.delete_confirm",
					)
			: z.string(),
	});

	const form = useForm<AccountStatusFormValues>({
		resolver: zodResolver(schema),
		defaultValues: { currentPassword: "", confirmText: "" },
	});

	const onSubmit = async (values: AccountStatusFormValues) => {
		if (ownedGroups.length > 0 || pending) {
			return;
		}
		const payload = { current_password: hasPassword ? values.currentPassword : "" };
		try {
			if (isDelete) {
				await deleteMutation.mutateAsync(payload);
				toast.success(t("user_settings.account.delete_success"));
			} else {
				await disableMutation.mutateAsync(payload);
				toast.success(t("user_settings.account.disable_success"));
			}
			onClose();
			onAccountClosed?.();
			router.replace("/sign-in");
		} catch (err) {
			applySettingsFormError(form, err, t, {
				passwordField: hasPassword ? "currentPassword" : undefined,
			});
		}
	};

	if (ownedGroups.length > 0) {
		return (
			<>
				<WiChatDialogHeader
					title={
						isDelete
							? t("user_settings.account.delete_dialog_title")
							: t("user_settings.account.disable_dialog_title")
					}
					subtitle={t("user_settings.account.owns_groups_alert_title")}
					titleId={titleId}
				/>
				<WiChatDialogBody className="pb-5">
					<div
						role="alert"
						className="flex items-start gap-2.5 rounded-xl border border-border bg-muted/40 p-3.5"
					>
						<AlertTriangle className="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden />
						<p className="text-xs leading-relaxed text-muted-foreground">
							{t("user_settings.account.owns_groups_alert_desc")}
						</p>
					</div>
				</WiChatDialogBody>
				<WiChatDialogFooter>
					<Button type="button" onClick={onClose} className={authSubmitClass}>
						{t("user_settings.account.owns_groups_close")}
					</Button>
				</WiChatDialogFooter>
			</>
		);
	}

	return (
		<>
			<WiChatDialogHeader
				title={
					isDelete
						? t("user_settings.account.delete_dialog_title")
						: t("user_settings.account.disable_dialog_title")
				}
				subtitle={
					isDelete
						? t("user_settings.account.delete_dialog_subtitle")
						: t("user_settings.account.disable_dialog_subtitle")
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
						{isDelete ? (
							<TextField
								control={form.control}
								name="confirmText"
								label={t("user_settings.account.delete_confirm_label")}
								placeholder={t("user_settings.account.delete_confirm_placeholder")}
								autoComplete="off"
								required
								labelClassName={authFieldLabelClass}
								inputClassName={authInputClass}
							/>
						) : null}
					</form>
				</Form>
			</WiChatDialogBody>
			<WiChatDialogFooter>
				<WiChatDialogFooterActions
					backLabel={t("user_settings.account.change_username_cancel")}
					onBack={onClose}
					primaryLabel={
						isDelete
							? t("user_settings.account.delete_confirm_action")
							: t("user_settings.account.disable_confirm_action")
					}
					primaryType="submit"
					primaryForm={formId}
					primaryDisabled={pending}
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

export function AccountStatusDialog({
	open,
	onOpenChange,
	mode,
	hasPassword,
	onAccountClosed,
}: AccountStatusDialogProps) {
	const titleId = useId();

	return (
		<WiChatDialog
			elevated
			open={open}
			onOpenChange={onOpenChange}
			titleId={titleId}
			className="max-w-[440px]"
		>
			<AccountStatusDialogBody
				mode={mode}
				hasPassword={hasPassword}
				titleId={titleId}
				onClose={() => onOpenChange(false)}
				onAccountClosed={onAccountClosed}
			/>
		</WiChatDialog>
	);
}
