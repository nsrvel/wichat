"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { AuthChallenge } from "@wichat/contracts/auth";
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
import { TurnstileWidget } from "@/features/auth/components/turnstile-widget";
import { authSubmitBlocked } from "@/features/auth/lib/captcha-policy";
import type { InstancePolicy } from "@wichat/contracts/auth";

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
	resendCooldownSec?: number;
	policy?: InstancePolicy;
	policyReady?: boolean;
	turnstileSiteKey?: string;
	turnstileToken?: string;
	onTurnstileToken?: (token: string) => void;
	turnstileResetKey?: number;
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
	resendCooldownSec = 0,
	policy,
	policyReady = true,
	turnstileSiteKey = "",
	turnstileToken = "",
	onTurnstileToken,
	turnstileResetKey = 0,
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
	const siteKey = turnstileSiteKey.trim();
	const submitBlocked = authSubmitBlocked(policy, policyReady, turnstileToken);

	return (
		<Form {...form}>
			<form
				className={authFormClass}
				onSubmit={form.handleSubmit((values) => onSubmit(values.code))}
			>
				<p className="text-center text-xs font-light text-muted-foreground sm:text-sm">
					{t(leadKey)}
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
				{siteKey ? (
					<TurnstileWidget
						siteKey={siteKey}
						resetKey={turnstileResetKey}
						onToken={(token) => onTurnstileToken?.(token)}
					/>
				) : null}
				<Button
					type="submit"
					disabled={isPending || submitBlocked}
					className={`mt-2 ${authSubmitClass}`}
				>
					{isPending ? t(submittingKey) : t(submitKey)}
				</Button>
			</form>
		</Form>
	);
}
