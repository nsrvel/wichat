"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { OtpField } from "@/components/form/otp-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import { authFormClass, authSubmitClass } from "@/lib/form-field-styles";
import { translateError } from "@/integrations/i18n/translate-error";
import { TurnstileWidget } from "@/features/auth/components/turnstile-widget";
import { ForgotPasswordResendCode } from "./forgot-password-resend-code";

const OTP_LENGTH = 6;

const ForgotPasswordOtpSchema = z.object({
	code: z.string().regex(new RegExp(`^[A-Z0-9]{${OTP_LENGTH}}$`), "validation.otp"),
});

export type ForgotPasswordOtpValues = z.infer<typeof ForgotPasswordOtpSchema>;

type ForgotPasswordOtpFormProps = {
	isPending?: boolean;
	error?: unknown;
	resendCooldownSec?: number;
	submitDisabled?: boolean;
	turnstileSiteKey?: string;
	turnstileResetKey?: number;
	onTurnstileToken?: (token: string) => void;
	onResend?: () => void;
	onSubmit: (values: ForgotPasswordOtpValues) => void;
};

export function ForgotPasswordOtpForm({
	isPending = false,
	error,
	resendCooldownSec,
	submitDisabled = false,
	turnstileSiteKey = "",
	turnstileResetKey = 0,
	onTurnstileToken,
	onResend,
	onSubmit,
}: ForgotPasswordOtpFormProps) {
	const { t } = useTranslation();
	const form = useForm<ForgotPasswordOtpValues>({
		resolver: zodResolver(ForgotPasswordOtpSchema),
		defaultValues: { code: "" },
	});

	return (
		<Form {...form}>
			<form className={authFormClass} onSubmit={form.handleSubmit(onSubmit)}>
				<OtpField
					control={form.control}
					name="code"
					label={t("forgot_password.otp_field_label")}
					length={OTP_LENGTH}
					required
					hideLabel
				/>
				{error ? (
					<p className="text-destructive text-[11px] font-medium">{translateError(t, error)}</p>
				) : null}
				{turnstileSiteKey ? (
					<TurnstileWidget
						siteKey={turnstileSiteKey}
						resetKey={turnstileResetKey}
						onVerify={(token) => onTurnstileToken?.(token)}
						onExpire={() => onTurnstileToken?.("")}
					/>
				) : null}
				<Button type="submit" disabled={isPending || submitDisabled} className={`mt-2 ${authSubmitClass}`}>
					{isPending ? t("forgot_password.confirming_code") : t("forgot_password.confirm_code")}
				</Button>
				<ForgotPasswordResendCode cooldownSec={resendCooldownSec} onResend={onResend} />
			</form>
		</Form>
	);
}
