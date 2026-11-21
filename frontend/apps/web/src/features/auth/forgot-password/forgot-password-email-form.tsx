"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import {
	authFieldLabelClass,
	authFormClass,
	authInputClass,
	authSubmitClass,
} from "@/lib/form-field-styles";
import { translateError } from "@/integrations/i18n/translate-error";
import { TurnstileWidget } from "@/features/auth/components/turnstile-widget";
import { ForgotPasswordBackToLogin } from "./forgot-password-back-to-login";

const ForgotPasswordEmailSchema = z.object({
	email: z.email("validation.email"),
});

export type ForgotPasswordEmailValues = z.infer<typeof ForgotPasswordEmailSchema>;

type ForgotPasswordEmailFormProps = {
	defaultEmail?: string;
	isPending?: boolean;
	error?: unknown;
	submitDisabled?: boolean;
	turnstileSiteKey?: string;
	turnstileResetKey?: number;
	onTurnstileToken?: (token: string) => void;
	onSubmit: (values: ForgotPasswordEmailValues) => void;
};

export function ForgotPasswordEmailForm({
	defaultEmail = "",
	isPending = false,
	error,
	submitDisabled = false,
	turnstileSiteKey = "",
	turnstileResetKey = 0,
	onTurnstileToken,
	onSubmit,
}: ForgotPasswordEmailFormProps) {
	const { t } = useTranslation();
	const form = useForm<ForgotPasswordEmailValues>({
		resolver: zodResolver(ForgotPasswordEmailSchema),
		defaultValues: { email: defaultEmail },
	});

	return (
		<Form {...form}>
			<form className={authFormClass} onSubmit={form.handleSubmit(onSubmit)}>
				<TextField
					control={form.control}
					name="email"
					type="email"
					label={t("login.email")}
					placeholder={t("login.email_placeholder")}
					autoComplete="email"
					required
					labelClassName={authFieldLabelClass}
					inputClassName={authInputClass}
				/>
				{error ? (
					<p className="text-destructive text-[11px] font-medium">{translateError(t, error)}</p>
				) : null}
				{turnstileSiteKey ? (
					<TurnstileWidget
						siteKey={turnstileSiteKey}
						resetKey={turnstileResetKey}
						onToken={(token) => onTurnstileToken?.(token)}
					/>
				) : null}
				<Button
					type="submit"
					disabled={isPending || submitDisabled}
					className={`mt-2 ${authSubmitClass}`}
				>
					{isPending ? t("forgot_password.sending_code") : t("forgot_password.send_code")}
				</Button>
				<ForgotPasswordBackToLogin />
			</form>
		</Form>
	);
}
