"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { AuthChallenge } from "@wichat/contracts";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { OtpField } from "@/components/form/otp-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import { authFormClass, authSubmitClass } from "@/features/auth/components/auth-shell";
import { translateError } from "@/integrations/i18n/translate-error";
import { ForgotPasswordResendCode } from "@/features/auth/forgot-password/forgot-password-resend-code";

const OTP_LENGTH = 6;

const OtpSchema = z.object({
	code: z.string().regex(new RegExp(`^[A-Z0-9]{${OTP_LENGTH}}$`), "validation.otp"),
});

type OtpValues = z.infer<typeof OtpSchema>;

type AuthEmailOtpStepProps = {
	challenge: AuthChallenge;
	leadKey: string;
	submitKey: string;
	submittingKey: string;
	isPending?: boolean;
	error?: unknown;
	onSubmit: (code: string) => void;
	onResend: () => void;
	resendCooldownSec?: number;
};

export function AuthEmailOtpStep({
	challenge,
	leadKey,
	submitKey,
	submittingKey,
	isPending = false,
	error,
	onSubmit,
	onResend,
	resendCooldownSec,
}: AuthEmailOtpStepProps) {
	const { t } = useTranslation();
	const form = useForm<OtpValues>({
		resolver: zodResolver(OtpSchema),
		defaultValues: { code: "" },
	});

	return (
		<Form {...form}>
			<form
				className={authFormClass}
				onSubmit={form.handleSubmit((values) => onSubmit(values.code))}
			>
				<p className="text-center text-xs font-light text-[#6d6d6d] sm:text-sm dark:text-[#a3a3a3]">
					{t(leadKey, { email: challenge.email_hint })}
				</p>
				<OtpField
					control={form.control}
					name="code"
					length={OTP_LENGTH}
					label={t("forgot_password.otp_field_label")}
					required
					hideLabel
				/>
				<ForgotPasswordResendCode
					cooldownSec={resendCooldownSec ?? challenge.resend_after_seconds}
					onResend={onResend}
				/>
				{error ? (
					<p className="text-destructive text-[11px] font-medium">
						{translateError(t, error)}
					</p>
				) : null}
				<Button
					type="submit"
					disabled={isPending}
					className={`mt-2 ${authSubmitClass}`}
				>
					{isPending ? t(submittingKey) : t(submitKey)}
				</Button>
			</form>
		</Form>
	);
}
