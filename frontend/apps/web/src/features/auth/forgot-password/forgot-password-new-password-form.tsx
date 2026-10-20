"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { PasswordField } from "@/components/form/password-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import {
	authFieldLabelClass,
	authFormClass,
	authInputClass,
	authSubmitClass,
} from "@/features/auth/components/auth-shell";
import { translateError } from "@/integrations/i18n/translate-error";

const ForgotPasswordNewPasswordSchema = z
	.object({
		password: z.string().min(12, "validation.password").max(128),
		confirm_password: z.string().min(12, "validation.password").max(128),
	})
	.refine((values) => values.password === values.confirm_password, {
		message: "validation.password_mismatch",
		path: ["confirm_password"],
	});

export type ForgotPasswordNewPasswordValues = z.infer<typeof ForgotPasswordNewPasswordSchema>;

type ForgotPasswordNewPasswordFormProps = {
	isPending?: boolean;
	error?: unknown;
	onSubmit: (values: ForgotPasswordNewPasswordValues) => void;
};

export function ForgotPasswordNewPasswordForm({
	isPending = false,
	error,
	onSubmit,
}: ForgotPasswordNewPasswordFormProps) {
	const { t } = useTranslation();
	const form = useForm<ForgotPasswordNewPasswordValues>({
		resolver: zodResolver(ForgotPasswordNewPasswordSchema),
		defaultValues: { password: "", confirm_password: "" },
	});

	return (
		<Form {...form}>
			<form className={authFormClass} onSubmit={form.handleSubmit(onSubmit)}>
				<PasswordField
					control={form.control}
					name="password"
					label={t("forgot_password.new_password")}
					placeholder={t("register.password_placeholder")}
					autoComplete="new-password"
					required
					showLabel={t("login.show_password")}
					hideLabel={t("login.hide_password")}
					labelClassName={authFieldLabelClass}
					inputClassName={authInputClass}
				/>
				<PasswordField
					control={form.control}
					name="confirm_password"
					label={t("forgot_password.confirm_password")}
					placeholder={t("forgot_password.confirm_password_placeholder")}
					autoComplete="new-password"
					required
					showLabel={t("login.show_password")}
					hideLabel={t("login.hide_password")}
					labelClassName={authFieldLabelClass}
					inputClassName={authInputClass}
				/>
				{error ? (
					<p className="text-destructive text-[11px] font-medium">
						{translateError(t, error)}
					</p>
				) : null}
				<Button type="submit" disabled={isPending} className={`mt-2 ${authSubmitClass}`}>
					{isPending ? t("forgot_password.resetting") : t("forgot_password.reset_submit")}
				</Button>
			</form>
		</Form>
	);
}
