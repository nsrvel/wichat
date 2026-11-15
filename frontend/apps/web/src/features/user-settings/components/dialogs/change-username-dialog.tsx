"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { usernameSchema } from "@wichat/contracts/common";
import { Check, Loader2, X } from "lucide-react";
import { useEffect, useId, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";
import { PasswordField } from "@/components/form/password-field";
import { TextField } from "@/components/form/text-field";
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
import { checkUsername } from "@/services/auth/auth.service";
import { useUpdateMyUsernameMutation } from "@/services/user/user.mutation";

type UsernameCheckState =
	| { status: "idle" }
	| { status: "checking" }
	| { status: "available"; normalized: string }
	| { status: "taken" };

type ChangeUsernameDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	currentUsername: string;
};

const ChangeUsernameFormSchema = z.object({
	username: usernameSchema,
	currentPassword: z.string().min(1, "validation.password_required"),
});

type ChangeUsernameFormValues = z.infer<typeof ChangeUsernameFormSchema>;

function ChangeUsernameDialogBody({
	currentUsername,
	onClose,
	titleId,
}: {
	currentUsername: string;
	onClose: () => void;
	titleId: string;
}) {
	const { t } = useTranslation();
	const formId = `${titleId}-form`;
	const updateUsername = useUpdateMyUsernameMutation();

	const form = useForm<ChangeUsernameFormValues>({
		resolver: zodResolver(ChangeUsernameFormSchema),
		defaultValues: {
			username: currentUsername,
			currentPassword: "",
		},
	});
	const usernameDraft = useWatch({ control: form.control, name: "username" });
	const [usernameCheck, setUsernameCheck] = useState<{
		query: string;
		result: UsernameCheckState;
	} | null>(null);

	const normalizedCurrent = currentUsername.trim().toLowerCase();
	const trimmedUsername = usernameDraft.trim().toLowerCase();
	const isChanged = trimmedUsername !== normalizedCurrent;
	const isValidUsername = usernameSchema.safeParse(trimmedUsername).success;

	useEffect(() => {
		if (!isChanged || !isValidUsername) {
			return;
		}
		let cancelled = false;
		const timer = setTimeout(async () => {
			try {
				const res = await checkUsername(trimmedUsername);
				if (cancelled) return;
				setUsernameCheck({
					query: trimmedUsername,
					result: res.available
						? { status: "available", normalized: res.normalized }
						: { status: "taken" },
				});
			} catch {
				if (!cancelled) {
					setUsernameCheck({ query: trimmedUsername, result: { status: "idle" } });
				}
			}
		}, 350);
		return () => {
			cancelled = true;
			clearTimeout(timer);
		};
	}, [isChanged, isValidUsername, trimmedUsername]);

	const effectiveUsernameCheck: UsernameCheckState =
		!isChanged || !isValidUsername
			? { status: "idle" }
			: usernameCheck?.query === trimmedUsername
				? usernameCheck.result
				: { status: "checking" };

	const onSubmit = async (values: ChangeUsernameFormValues) => {
		if (
			!isChanged ||
			effectiveUsernameCheck.status === "taken" ||
			effectiveUsernameCheck.status === "checking" ||
			updateUsername.isPending
		) {
			return;
		}
		try {
			await updateUsername.mutateAsync({
				username: values.username.trim().toLowerCase(),
				current_password: values.currentPassword,
			});
			toast.success(t("user_settings.account.change_username_success"));
			onClose();
		} catch (err) {
			applySettingsFormError(form, err, t, {
				passwordField: "currentPassword",
				codeMap: { "user.username_taken": "username" },
				paramFieldMap: {
					username: "username",
					current_password: "currentPassword",
				},
			});
		}
	};

	return (
		<>
			<WiChatDialogHeader
				title={t("user_settings.account.change_username_title")}
				subtitle={t("user_settings.account.change_username_subtitle")}
				titleId={titleId}
				className="text-left"
			/>

			<WiChatDialogBody>
				<Form {...form}>
					<form id={formId} onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
						<div className="space-y-1.5">
							<TextField
								control={form.control}
								name="username"
								label={t("user_settings.account.username")}
								placeholder="username"
								autoComplete="username"
								maxLength={32}
								required
								labelClassName={authFieldLabelClass}
								inputClassName={authInputClass}
							/>
							{effectiveUsernameCheck.status === "checking" ? (
								<p className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
									<Loader2 className="size-3 shrink-0 animate-spin" aria-hidden />
									<span>{t("register.username_checking")}</span>
								</p>
							) : null}
							{effectiveUsernameCheck.status === "available" ? (
								<p className="flex items-center gap-1.5 text-[11px] font-medium text-emerald-500 dark:text-emerald-400">
									<Check className="size-3.5 shrink-0" aria-hidden />
									<span>
										{t("register.username_available", {
											username: effectiveUsernameCheck.normalized,
										})}
									</span>
								</p>
							) : null}
							{effectiveUsernameCheck.status === "taken" ? (
								<p className="flex items-center gap-1.5 text-[11px] font-medium text-destructive">
									<X className="size-3.5 shrink-0" aria-hidden />
									<span>{t("register.username_taken")}</span>
								</p>
							) : null}
						</div>

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
				</Form>
			</WiChatDialogBody>

			<WiChatDialogFooter>
				<WiChatDialogFooterActions
					backLabel={t("user_settings.account.change_username_cancel")}
					onBack={onClose}
					primaryLabel={t("user_settings.account.change_username_done")}
					primaryType="submit"
					primaryForm={formId}
					primaryDisabled={
						!isChanged ||
						effectiveUsernameCheck.status === "taken" ||
						effectiveUsernameCheck.status === "checking" ||
						updateUsername.isPending
					}
					backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
					primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
				/>
			</WiChatDialogFooter>
		</>
	);
}

export function ChangeUsernameDialog({
	open,
	onOpenChange,
	currentUsername,
}: ChangeUsernameDialogProps) {
	const titleId = useId();

	return (
		<WiChatDialog
			elevated
			open={open}
			onOpenChange={onOpenChange}
			titleId={titleId}
			className="max-w-[440px]"
		>
			<ChangeUsernameDialogBody
				currentUsername={currentUsername}
				titleId={titleId}
				onClose={() => onOpenChange(false)}
			/>
		</WiChatDialog>
	);
}
