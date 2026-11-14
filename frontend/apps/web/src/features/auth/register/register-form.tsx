"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import {
	displayNameSchema,
	usernameSchema,
	type AuthChallenge,
	type RegisterRequest,
} from "@wichat/contracts";
import { Check, Loader2, X } from "lucide-react";
import { useEffect, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { BirthDateField } from "@/components/form/birth-date-field";
import { PasswordField } from "@/components/form/password-field";
import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import { AuthEmailOtpStep } from "@/features/auth/components/auth-email-otp-step";
import { TurnstileWidget } from "@/features/auth/components/turnstile-widget";
import { authFieldLabelClass, authFormClass, authInputClass, authSubmitClass } from "@/lib/form-field-styles";
import { resendRateLimitCooldownSec } from "@/features/auth/lib/resend-cooldown";
import { translateError } from "@/integrations/i18n/translate-error";
import { birthPartsDefaults, formatBirthDate, refineMinAgeBirth } from "@/lib/auth-validators";
import {
	useRegisterMutation,
	useResendRegisterCodeMutation,
	useVerifyRegisterMutation,
} from "@/services/auth/auth.mutation";
import { instancePolicyQueryOpts } from "@/services/auth/auth.query";
import { checkUsername } from "@/services/auth/auth.service";

const birthDaySchema = z.string().regex(/^\d{2}$/);
const birthMonthSchema = z.string().regex(/^\d{2}$/);
const birthYearSchema = z.string().regex(/^\d{4}$/);

const RegisterFormSchema = z
	.object({
		email: z.email("validation.email"),
		username: usernameSchema,
		password: z.string().min(12, "validation.password").max(128),
		display_name: displayNameSchema,
		birth_day: birthDaySchema,
		birth_month: birthMonthSchema,
		birth_year: birthYearSchema,
	})
	.superRefine(refineMinAgeBirth);

type RegisterFormValues = z.infer<typeof RegisterFormSchema>;

function toRegisterRequest(
	values: RegisterFormValues,
	locale: string,
	turnstileToken: string,
): RegisterRequest {
	return {
		email: values.email,
		username: values.username,
		password: values.password,
		display_name: values.display_name,
		birth_date: formatBirthDate(values.birth_year, values.birth_month, values.birth_day),
		locale,
		turnstile_token: turnstileToken || undefined,
	};
}

type UsernameCheckState =
	| { status: "idle" }
	| { status: "checking" }
	| { status: "available"; normalized: string }
	| { status: "taken" };

export function RegisterForm() {
	const { t, i18n } = useTranslation();
	const register = useRegisterMutation();
	const verify = useVerifyRegisterMutation();
	const resend = useResendRegisterCodeMutation();
	const { data: policy } = useQuery(instancePolicyQueryOpts());
	const turnstileSiteKey = policy?.turnstile_site_key ?? "";
	const [challenge, setChallenge] = useState<AuthChallenge | null>(null);
	const [resendCooldown, setResendCooldown] = useState(0);
	const [turnstileToken, setTurnstileToken] = useState("");
	const [turnstileResetKey, setTurnstileResetKey] = useState(0);
	const [usernameCheck, setUsernameCheck] = useState<
		{ query: string; result: UsernameCheckState } | null
	>(null);
	const form = useForm<RegisterFormValues>({
		resolver: zodResolver(RegisterFormSchema),
		defaultValues: {
			email: "",
			username: "",
			password: "",
			display_name: "",
			...birthPartsDefaults(),
		},
	});

	const watchedUsername = useWatch({ control: form.control, name: "username" }) ?? "";
	const trimmedUsername = watchedUsername.trim().toLowerCase();
	const isValidUsername = usernameSchema.safeParse(trimmedUsername).success;

	useEffect(() => {
		if (!isValidUsername) {
			return;
		}
		let cancelled = false;
		const timer = setTimeout(async () => {
			try {
				const res = await checkUsername(trimmedUsername);
				if (cancelled) return;
				setUsernameCheck({
					query: trimmedUsername,
					result: res.available
						? { status: "available", normalized: res.normalized }
						: { status: "taken" },
				});
			} catch {
				if (!cancelled) {
					setUsernameCheck({ query: trimmedUsername, result: { status: "idle" } });
				}
			}
		}, 350);
		return () => {
			cancelled = true;
			clearTimeout(timer);
		};
	}, [isValidUsername, trimmedUsername]);

	const effectiveUsernameCheck: UsernameCheckState = !isValidUsername
		? { status: "idle" }
		: usernameCheck?.query === trimmedUsername
			? usernameCheck.result
			: { status: "checking" };

	if (challenge) {
		return (
			<AuthEmailOtpStep
				challenge={challenge}
				leadKey="register.code_lead"
				submitKey="register.confirm_code"
				submittingKey="register.confirming_code"
				isPending={verify.isPending}
				error={verify.error}
				resendError={resend.error}
				isResendPending={resend.isPending}
				resendCooldownSec={resendCooldown}
				onResend={() => {
					verify.reset();
					resend.mutate(
						{ challenge_id: challenge.challenge_id },
						{
							onSuccess: (data) => setResendCooldown(data.resend_after_seconds),
							onError: (err) => {
								const wait = resendRateLimitCooldownSec(err);
								if (wait !== null) {
									setResendCooldown(wait);
								}
							},
						},
					);
				}}
				onBack={() => {
					verify.reset();
					resend.reset();
					register.reset();
					setChallenge(null);
					setResendCooldown(0);
				}}
				onSubmit={(code) => {
					resend.reset();
					verify.mutate({ challenge_id: challenge.challenge_id, code });
				}}
			/>
		);
	}

	return (
		<Form {...form}>
			<form
				className={authFormClass}
				onSubmit={form.handleSubmit((values) => {
					register.mutate(toRegisterRequest(values, i18n.language, turnstileToken), {
						onSuccess: (step) => {
							if (step.kind === "challenge") {
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
					name="email"
					type="email"
					label={t("login.email")}
					placeholder={t("register.email_placeholder")}
					autoComplete="email"
					required
					labelClassName={authFieldLabelClass}
					inputClassName={authInputClass}
				/>
				<div className="space-y-1">
					<TextField
						control={form.control}
						name="username"
						label={t("register.username")}
						placeholder={t("register.username_placeholder")}
						autoComplete="username"
						required
						labelClassName={authFieldLabelClass}
						inputClassName={authInputClass}
					/>
					{effectiveUsernameCheck.status === "checking" ? (
						<p className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
							<Loader2 className="size-3 shrink-0 animate-spin" aria-hidden />
							<span>{t("register.username_checking")}</span>
						</p>
					) : null}
					{effectiveUsernameCheck.status === "available" ? (
						<p className="flex items-center gap-1.5 text-[11px] font-medium text-emerald-500 dark:text-emerald-400">
							<Check className="size-3.5 shrink-0" aria-hidden />
							<span>
								{t("register.username_available", { username: effectiveUsernameCheck.normalized })}
							</span>
						</p>
					) : null}
					{effectiveUsernameCheck.status === "taken" ? (
						<p className="flex items-center gap-1.5 text-[11px] font-medium text-destructive">
							<X className="size-3.5 shrink-0" aria-hidden />
							<span>{t("register.username_taken")}</span>
						</p>
					) : null}
				</div>
				<TextField
					control={form.control}
					name="display_name"
					label={t("register.display_name")}
					placeholder={t("register.display_placeholder")}
					autoComplete="name"
					required
					labelClassName={authFieldLabelClass}
					inputClassName={authInputClass}
				/>
				<PasswordField
					control={form.control}
					name="password"
					label={t("login.password")}
					placeholder={t("register.password_placeholder")}
					autoComplete="new-password"
					required
					showLabel={t("login.show_password")}
					hideLabel={t("login.hide_password")}
					labelClassName={authFieldLabelClass}
					inputClassName={authInputClass}
				/>
				<BirthDateField
					control={form.control}
					dayName="birth_day"
					monthName="birth_month"
					yearName="birth_year"
					label={t("register.birth_date")}
					dayPlaceholder={t("register.dob_day")}
					monthPlaceholder={t("register.dob_month")}
					yearPlaceholder={t("register.dob_year")}
					required
					labelClassName={authFieldLabelClass}
					selectClassName={authInputClass}
				/>
				<TurnstileWidget
					siteKey={turnstileSiteKey}
					resetKey={turnstileResetKey}
					onToken={setTurnstileToken}
				/>
				{register.error ? (
					<p role="alert" className="text-destructive text-[11px] font-medium">
						{translateError(t, register.error)}
					</p>
				) : null}
				<Button
					type="submit"
					disabled={
						register.isPending ||
						effectiveUsernameCheck.status === "checking" ||
						effectiveUsernameCheck.status === "taken" ||
						Boolean(turnstileSiteKey && !turnstileToken)
					}
					className={turnstileSiteKey ? authSubmitClass : `mt-2 ${authSubmitClass}`}
				>
					{register.isPending ? t("register.submitting") : t("register.submit")}
				</Button>
			</form>
		</Form>
	);
}
