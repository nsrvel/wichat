"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { AuthChallenge } from "@wichat/contracts";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { OtpField } from "@/components/form/otp-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import { authFormClass, authSubmitClass } from "@/lib/form-field-styles";
import { ApiError } from "@wichat/api-client/api-error";
import { translateError } from "@/integrations/i18n/translate-error";
import { effectiveResendCooldownSec } from "@/features/auth/lib/resend-cooldown";
import { ForgotPasswordResendCode } from "@/features/auth/forgot-password/forgot-password-resend-code";

const OTP_LENGTH = 6;

const OtpSchema = z.object({
	code: z.string().regex(new RegExp(`^[A-Z0-9]{${OTP_LENGTH}}$`), "validation.otp"),
});

type OtpValues = z.infer<typeof OtpSchema>;

function sameApiError(a: unknown, b: unknown): boolean {
	if (!(a instanceof ApiError) || !(b instanceof ApiError)) {
		return false;
	}
	if (a.code !== b.code) {
		return false;
	}
	return JSON.stringify(a.params ?? null) === JSON.stringify(b.params ?? null);
}

type AuthEmailOtpStepProps = {
	challenge: AuthChallenge;
	leadKey: string;
	submitKey: string;
	submittingKey: string;
	isPending?: boolean;
	error?: unknown;
	resendError?: unknown;
	isResendPending?: boolean;
	onSubmit: (code: string) => void;
	onResend: () => void;
	onBack?: () => void;
	backLabel?: string;
	resendCooldownSec?: number;
};

export function AuthEmailOtpStep({
	challenge,
	leadKey,
	submitKey,
	submittingKey,
	isPending = false,
	error,
	resendError,
	isResendPending = false,
	onSubmit,
	onResend,
	onBack,
	backLabel,
	resendCooldownSec = 0,
}: AuthEmailOtpStepProps) {
	const { t } = useTranslation();
	const resendCooldown = effectiveResendCooldownSec(
		resendCooldownSec,
		challenge.resend_after_seconds,
	);
	const form = useForm<OtpValues>({
		resolver: zodResolver(OtpSchema),
		defaultValues: { code: "" },
	});
	const showVerifyError = error != null && !sameApiError(error, resendError);

	return (
		<Form {...form}>
			<form
				className={authFormClass}
				onSubmit={form.handleSubmit((values) => onSubmit(values.code))}
			>
				<p className="text-center text-xs font-light text-muted-foreground sm:text-sm">
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
					cooldownSec={resendCooldown}
					onResend={onResend}
					isResending={isResendPending}
				/>
				{resendError ? (
					<p className="text-destructive text-[11px] font-medium">
						{translateError(t, resendError)}
					</p>
				) : null}
				{showVerifyError ? (
					<p className="text-destructive text-[11px] font-medium">{translateError(t, error)}</p>
				) : null}
				<Button type="submit" disabled={isPending} className={`mt-2 ${authSubmitClass}`}>
					{isPending ? t(submittingKey) : t(submitKey)}
				</Button>
				{onBack ? (
					<Button
						type="button"
						variant="ghost"
						disabled={isPending}
						onClick={onBack}
						className="h-9 w-full text-xs font-medium text-muted-foreground hover:text-foreground"
					>
						{backLabel ?? t("auth.back")}
					</Button>
				) : null}
			</form>
		</Form>
	);
}
