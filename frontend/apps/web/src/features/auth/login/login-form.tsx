"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import type { AuthChallenge, LoginRequest } from "@wichat/contracts/auth";
import Link from "next/link";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { PasswordField } from "@/components/form/password-field";
import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import { AuthEmailOtpStep } from "@/features/auth/components/auth-email-otp-step";
import { AuthMfaStep } from "@/features/auth/components/auth-mfa-step";
import { AuthPrivacyNotice } from "@/features/auth/components/auth-privacy-notice";
import { AuthOr } from "@/features/auth/components/auth-shell";
import { GoogleButton } from "@/features/auth/components/google-button";
import { TurnstileWidget } from "@/features/auth/components/turnstile-widget";
import {
	authFieldLabelClass,
	authFormClass,
	authInputClass,
	authFieldLinkClass,
	authSubmitClass,
} from "@/lib/form-field-styles";
import { resendRateLimitCooldownSec } from "@/features/auth/lib/resend-cooldown";
import { translateError } from "@/integrations/i18n/translate-error";
import {
	useLoginMutation,
	useResendLoginCodeMutation,
	useVerifyLoginMutation,
} from "@/services/auth/auth.mutation";
import { instancePolicyQueryOpts } from "@/services/auth/auth.query";
import { authSubmitBlocked } from "@/features/auth/lib/captcha-policy";

const LoginFormSchema = z.object({
	login: z.string().trim().min(1, "validation.login").max(254),
	password: z.string().min(1, "validation.password_required").max(128),
});

type LoginFormValues = z.infer<typeof LoginFormSchema>;

export function LoginForm({ authReturnTo = "/" }: { authReturnTo?: string }) {
	const { t, i18n } = useTranslation();
	const login = useLoginMutation();
	const verify = useVerifyLoginMutation();
	const resend = useResendLoginCodeMutation();
	const { data: policy, isSuccess: policyReady } = useQuery(instancePolicyQueryOpts());
	const turnstileSiteKey = policy?.turnstile_site_key ?? "";
	const [challenge, setChallenge] = useState<AuthChallenge | null>(null);
	const [mfaChallenge, setMfaChallenge] = useState<AuthChallenge | null>(null);
	const [resendCooldown, setResendCooldown] = useState(0);
	const [turnstileToken, setTurnstileToken] = useState("");
	const [turnstileResetKey, setTurnstileResetKey] = useState(0);
	const [otpTurnstileToken, setOtpTurnstileToken] = useState("");
	const [otpTurnstileResetKey, setOtpTurnstileResetKey] = useState(0);
	const credentialBlocked = authSubmitBlocked(policy, policyReady, turnstileToken);
	const form = useForm<LoginFormValues>({
		resolver: zodResolver(LoginFormSchema),
		defaultValues: { login: "", password: "" },
	});

	if (mfaChallenge) {
		return (
			<AuthMfaStep
				challenge={mfaChallenge}
				onSuccess={() => {
					setMfaChallenge(null);
				}}
				onBack={() => {
					login.reset();
					setMfaChallenge(null);
				}}
			/>
		);
	}

	if (challenge) {
		return (
			<AuthEmailOtpStep
				challenge={challenge}
				policy={policy}
				policyReady={policyReady}
				turnstileSiteKey={turnstileSiteKey}
				turnstileToken={otpTurnstileToken}
				onTurnstileToken={setOtpTurnstileToken}
				turnstileResetKey={otpTurnstileResetKey}
				leadKey="login.code_lead_generic"
				submitKey="login.confirm_code"
				submittingKey="login.confirming_code"
				isPending={verify.isPending}
				error={verify.error}
				resendError={resend.error}
				isResendPending={resend.isPending}
				resendCooldownSec={resendCooldown}
				onResend={() => {
					verify.reset();
					resend.mutate(
						{
							challenge_id: challenge.challenge_id,
							turnstile_token: otpTurnstileToken || undefined,
						},
						{
							onSuccess: (data) => {
								setResendCooldown(data.resend_after_seconds);
								setOtpTurnstileToken("");
								setOtpTurnstileResetKey((k) => k + 1);
							},
							onError: (err) => {
								setOtpTurnstileToken("");
								setOtpTurnstileResetKey((k) => k + 1);
								const wait = resendRateLimitCooldownSec(err);
								if (wait !== null) {
									setResendCooldown(wait);
								}
							},
						},
					);
				}}
				onSubmit={(code) => {
					resend.reset();
					verify.mutate(
						{
							challenge_id: challenge.challenge_id,
							code,
							turnstile_token: otpTurnstileToken || undefined,
						},
						{
							onError: () => {
								if (turnstileSiteKey) {
									setOtpTurnstileToken("");
									setOtpTurnstileResetKey((k) => k + 1);
								}
							},
						},
					);
				}}
			/>
		);
	}

	return (
		<>
			<GoogleButton
				label={t("auth.google")}
				authReturnTo={authReturnTo}
				turnstileToken={turnstileToken}
				disabled={credentialBlocked}
			/>
			<AuthOr label={t("auth.or")} />
			<Form {...form}>
				<form
					className={authFormClass}
					onSubmit={form.handleSubmit((values) => {
						if (credentialBlocked) {
							return;
						}
						const payload: LoginRequest = {
							...values,
							locale: i18n.language,
							turnstile_token: turnstileToken || undefined,
						};
						login.mutate(payload, {
							onSuccess: (step) => {
								if (step.kind === "mfa_challenge") {
									setMfaChallenge(step.challenge);
								} else if (step.kind === "challenge") {
									setChallenge(step.challenge);
									setResendCooldown(step.challenge.resend_after_seconds);
								}
							},
							onError: () => {
								if (turnstileSiteKey) {
									setTurnstileToken("");
									setTurnstileResetKey((k) => k + 1);
								}
							},
						});
					})}
				>
					<TextField
						control={form.control}
						name="login"
						label={t("login.login_id")}
						placeholder={t("login.login_placeholder")}
						autoComplete="username"
						required
						labelClassName={authFieldLabelClass}
						inputClassName={authInputClass}
					/>
					<PasswordField
						control={form.control}
						name="password"
						label={t("login.password")}
						autoComplete="current-password"
						required
						showLabel={t("login.show_password")}
						hideLabel={t("login.hide_password")}
						labelClassName={authFieldLabelClass}
						inputClassName={authInputClass}
						labelAction={
							<Link href="/forgot-password" className={authFieldLinkClass}>
								{t("login.forgot_password")}
							</Link>
						}
					/>
					<TurnstileWidget
						siteKey={turnstileSiteKey}
						resetKey={turnstileResetKey}
						onToken={setTurnstileToken}
					/>
					{login.error ? (
						<p role="alert" className="text-destructive text-[11px] font-medium">
							{translateError(t, login.error)}
						</p>
					) : null}
					<Button
						type="submit"
						disabled={login.isPending || credentialBlocked}
						className={turnstileSiteKey ? authSubmitClass : `mt-2 ${authSubmitClass}`}
					>
						{login.isPending ? t("login.submitting") : t("login.submit")}
					</Button>
				</form>
				<AuthPrivacyNotice />
			</Form>
		</>
	);
}
