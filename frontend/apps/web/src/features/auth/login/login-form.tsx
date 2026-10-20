"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { AuthChallenge, LoginRequest } from "@wichat/contracts";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { PasswordField } from "@/components/form/password-field";
import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import { AuthEmailOtpStep } from "@/features/auth/components/auth-email-otp-step";
import {
	authFieldLabelClass,
	authFormClass,
	authInputClass,
	authFieldLinkClass,
	authSubmitClass,
} from "@/features/auth/components/auth-shell";
import { translateError } from "@/integrations/i18n/translate-error";
import {
	useLoginMutation,
	useResendLoginCodeMutation,
	useVerifyLoginMutation,
} from "@/services/auth/auth.mutation";

const LoginFormSchema = z.object({
	login: z.string().trim().min(1, "validation.login").max(254),
	password: z.string().min(12, "validation.password").max(128),
});

type LoginFormValues = z.infer<typeof LoginFormSchema>;

export function LoginForm() {
	const { t, i18n } = useTranslation();
	const router = useRouter();
	const login = useLoginMutation();
	const verify = useVerifyLoginMutation();
	const resend = useResendLoginCodeMutation();
	const [challenge, setChallenge] = useState<AuthChallenge | null>(null);
	const [resendCooldown, setResendCooldown] = useState(0);
	const form = useForm<LoginFormValues>({
		resolver: zodResolver(LoginFormSchema),
		defaultValues: { login: "", password: "" },
	});

	if (challenge) {
		return (
			<AuthEmailOtpStep
				challenge={challenge}
				leadKey="login.code_lead"
				submitKey="login.confirm_code"
				submittingKey="login.confirming_code"
				isPending={verify.isPending}
				error={verify.error}
				resendCooldownSec={resendCooldown}
				onResend={() => {
					resend.mutate(
						{ challenge_id: challenge.challenge_id },
						{
							onSuccess: (data) => setResendCooldown(data.resend_after_seconds),
						},
					);
				}}
				onSubmit={(code) => {
					verify.mutate(
						{ challenge_id: challenge.challenge_id, code },
						{ onSuccess: () => router.push("/") },
					);
				}}
			/>
		);
	}

	return (
		<Form {...form}>
			<form
				className={authFormClass}
				onSubmit={form.handleSubmit((values) => {
					const payload: LoginRequest = {
						...values,
						locale: i18n.language,
					};
					login.mutate(payload, {
						onSuccess: (step) => {
							if (step.kind === "challenge") {
								setChallenge(step.challenge);
								setResendCooldown(step.challenge.resend_after_seconds);
								return;
							}
							router.push("/");
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
					placeholder={t("login.password_placeholder")}
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
				{login.error ? (
					<p className="text-destructive text-[11px] font-medium">
						{translateError(t, login.error)}
					</p>
				) : null}
				<Button type="submit" disabled={login.isPending} className={`mt-2 ${authSubmitClass}`}>
					{login.isPending ? t("login.submitting") : t("login.submit")}
				</Button>
			</form>
		</Form>
	);
}
