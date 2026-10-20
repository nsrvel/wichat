"use client";

import { ApiError } from "@wichat/api-client/api-error";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { AuthSessionLayout } from "@/features/auth/components/auth-session-layout";
import {
	useForgotPasswordRequestMutation,
	useForgotPasswordResetMutation,
	useForgotPasswordVerifyMutation,
} from "@/services/auth/auth.mutation";
import { ForgotPasswordEmailForm } from "./forgot-password-email-form";
import { ForgotPasswordNewPasswordForm } from "./forgot-password-new-password-form";
import { ForgotPasswordOtpForm } from "./forgot-password-otp-form";

type ForgotPasswordStep = "email" | "code" | "password";

export function ForgotPasswordPage({ notice }: { notice?: string }) {
	const { t, i18n } = useTranslation();
	const router = useRouter();
	const [step, setStep] = useState<ForgotPasswordStep>("email");
	const [email, setEmail] = useState("");
	const [resetToken, setResetToken] = useState("");
	const [resendCooldownSec, setResendCooldownSec] = useState(60);

	const requestReset = useForgotPasswordRequestMutation();
	const verifyReset = useForgotPasswordVerifyMutation();
	const completeReset = useForgotPasswordResetMutation();

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
		if (error instanceof ApiError) {
			const raw = error.params?.retry_after_seconds;
			if (typeof raw === "number" && raw > 0) {
				setResendCooldownSec(raw);
			}
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
					onSubmit={(values) => {
						requestReset.mutate(
							{ email: values.email, locale },
							{
								onSuccess: (data) => {
									setEmail(values.email);
									applyResendCooldown(data.resend_cooldown_seconds);
									setStep("code");
								},
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
					onResend={() => {
						requestReset.mutate(
							{ email, locale },
							{
								onSuccess: (data) => {
									applyResendCooldown(data.resend_cooldown_seconds);
								},
								onError: applyResendRateLimit,
							},
						);
					}}
					onSubmit={(values) => {
						verifyReset.mutate(
							{ email, code: values.code },
							{
								onSuccess: (data) => {
									setResetToken(data.reset_token);
									setStep("password");
								},
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
