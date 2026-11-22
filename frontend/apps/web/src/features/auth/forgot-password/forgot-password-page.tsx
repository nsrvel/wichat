"use client";

import { useQuery } from "@tanstack/react-query";
import { resendRateLimitCooldownSec } from "@/features/auth/lib/resend-cooldown";
import { authSubmitBlocked } from "@/features/auth/lib/captcha-policy";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { AuthSessionLayout } from "@/features/auth/components/auth-session-layout";
import {
	useForgotPasswordRequestMutation,
	useForgotPasswordResetMutation,
	useForgotPasswordVerifyMutation,
} from "@/services/auth/auth.mutation";
import { instancePolicyQueryOpts } from "@/services/auth/auth.query";
import { ForgotPasswordEmailForm } from "./forgot-password-email-form";
import { ForgotPasswordNewPasswordForm } from "./forgot-password-new-password-form";
import { ForgotPasswordOtpForm } from "./forgot-password-otp-form";

type ForgotPasswordStep = "email" | "code" | "password";

export function ForgotPasswordPage({ notice }: { notice?: string }) {
	const { t, i18n } = useTranslation();
	const router = useRouter();
	const { data: policy, isSuccess: policyReady } = useQuery(instancePolicyQueryOpts());
	const turnstileSiteKey = policy?.turnstile_site_key ?? "";
	const [step, setStep] = useState<ForgotPasswordStep>("email");
	const [email, setEmail] = useState("");
	const [resetToken, setResetToken] = useState("");
	const [resendCooldownSec, setResendCooldownSec] = useState(60);
	const [turnstileToken, setTurnstileToken] = useState("");
	const [turnstileResetKey, setTurnstileResetKey] = useState(0);
	const [otpTurnstileToken, setOtpTurnstileToken] = useState("");
	const [otpTurnstileResetKey, setOtpTurnstileResetKey] = useState(0);

	const requestReset = useForgotPasswordRequestMutation();
	const verifyReset = useForgotPasswordVerifyMutation();
	const completeReset = useForgotPasswordResetMutation();

	const emailSubmitBlocked = authSubmitBlocked(policy, policyReady, turnstileToken);
	const otpSubmitBlocked = authSubmitBlocked(policy, policyReady, otpTurnstileToken);

	const headline =
		step === "email"
			? t("forgot_password.title_email")
			: step === "code"
				? t("forgot_password.title_code")
				: t("forgot_password.title_password");

	const lead =
		step === "email"
			? t("forgot_password.lead_email")
			: step === "code"
				? t("forgot_password.lead_code", { email })
				: t("forgot_password.lead_password");

	const locale = i18n.language?.startsWith("id") ? "id" : "en";

	const applyResendCooldown = (seconds: number) => {
		if (seconds > 0) {
			setResendCooldownSec(seconds);
		}
	};

	const applyResendRateLimit = (error: unknown) => {
		const wait = resendRateLimitCooldownSec(error);
		if (wait !== null) {
			setResendCooldownSec(wait);
		}
	};

	const bumpEmailTurnstile = () => {
		if (turnstileSiteKey) {
			setTurnstileToken("");
			setTurnstileResetKey((k) => k + 1);
		}
	};

	const bumpOtpTurnstile = () => {
		if (turnstileSiteKey) {
			setOtpTurnstileToken("");
			setOtpTurnstileResetKey((k) => k + 1);
		}
	};

	return (
		<AuthSessionLayout
			headline={headline}
			loadingHeadline={t("forgot_password.title_email")}
			lead={lead}
			notice={notice}
		>
			{step === "email" ? (
				<ForgotPasswordEmailForm
					defaultEmail={email}
					isPending={requestReset.isPending}
					error={requestReset.error}
					submitDisabled={emailSubmitBlocked}
					turnstileSiteKey={turnstileSiteKey}
					turnstileResetKey={turnstileResetKey}
					onTurnstileToken={setTurnstileToken}
					onSubmit={(values) => {
						requestReset.mutate(
							{
								email: values.email,
								locale,
								turnstile_token: turnstileToken || undefined,
							},
							{
								onSuccess: (data) => {
									setEmail(values.email);
									applyResendCooldown(data.resend_cooldown_seconds);
									bumpOtpTurnstile();
									setStep("code");
								},
								onError: () => bumpEmailTurnstile(),
							},
						);
					}}
				/>
			) : null}
			{step === "code" ? (
				<ForgotPasswordOtpForm
					isPending={verifyReset.isPending}
					error={verifyReset.error ?? requestReset.error}
					resendCooldownSec={resendCooldownSec}
					submitDisabled={otpSubmitBlocked}
					turnstileSiteKey={turnstileSiteKey}
					turnstileResetKey={otpTurnstileResetKey}
					onTurnstileToken={setOtpTurnstileToken}
					onResend={() => {
						requestReset.mutate(
							{
								email,
								locale,
								turnstile_token: otpTurnstileToken || undefined,
							},
							{
								onSuccess: (data) => {
									applyResendCooldown(data.resend_cooldown_seconds);
									bumpOtpTurnstile();
								},
								onError: (err) => {
									applyResendRateLimit(err);
									bumpOtpTurnstile();
								},
							},
						);
					}}
					onSubmit={(values) => {
						verifyReset.mutate(
							{
								email,
								code: values.code,
								turnstile_token: otpTurnstileToken || undefined,
							},
							{
								onSuccess: (data) => {
									setResetToken(data.reset_token);
									setStep("password");
								},
								onError: () => bumpOtpTurnstile(),
							},
						);
					}}
				/>
			) : null}
			{step === "password" ? (
				<ForgotPasswordNewPasswordForm
					isPending={completeReset.isPending}
					error={completeReset.error}
					onSubmit={(values) => {
						completeReset.mutate(
							{ reset_token: resetToken, password: values.password },
							{
								onSuccess: () => {
									router.push("/sign-in");
								},
							},
						);
					}}
				/>
			) : null}
		</AuthSessionLayout>
	);
}
