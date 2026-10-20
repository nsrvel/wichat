"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { displayNameSchema, type AuthChallenge, type RegisterRequest } from "@wichat/contracts";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { BirthDateField } from "@/components/form/birth-date-field";
import { PasswordField } from "@/components/form/password-field";
import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import { AuthEmailOtpStep } from "@/features/auth/components/auth-email-otp-step";
import {
	authFieldLabelClass,
	authFormClass,
	authInputClass,
	authSubmitClass,
} from "@/features/auth/components/auth-shell";
import { translateError } from "@/integrations/i18n/translate-error";
import { birthPartsDefaults, formatBirthDate, refineMinAgeBirth } from "@/lib/auth-validators";
import {
	useRegisterMutation,
	useResendRegisterCodeMutation,
	useVerifyRegisterMutation,
} from "@/services/auth/auth.mutation";

const birthDaySchema = z.string().regex(/^\d{2}$/);
const birthMonthSchema = z.string().regex(/^\d{2}$/);
const birthYearSchema = z.string().regex(/^\d{4}$/);

const RegisterFormSchema = z
	.object({
		email: z.email("validation.email"),
		password: z.string().min(12, "validation.password").max(128),
		display_name: displayNameSchema,
		birth_day: birthDaySchema,
		birth_month: birthMonthSchema,
		birth_year: birthYearSchema,
	})
	.superRefine(refineMinAgeBirth);

type RegisterFormValues = z.infer<typeof RegisterFormSchema>;

function toRegisterRequest(values: RegisterFormValues, locale: string): RegisterRequest {
	return {
		email: values.email,
		password: values.password,
		display_name: values.display_name,
		birth_date: formatBirthDate(values.birth_year, values.birth_month, values.birth_day),
		locale,
	};
}

export function RegisterForm() {
	const { t, i18n } = useTranslation();
	const router = useRouter();
	const register = useRegisterMutation();
	const verify = useVerifyRegisterMutation();
	const resend = useResendRegisterCodeMutation();
	const [challenge, setChallenge] = useState<AuthChallenge | null>(null);
	const [resendCooldown, setResendCooldown] = useState(0);
	const form = useForm<RegisterFormValues>({
		resolver: zodResolver(RegisterFormSchema),
		defaultValues: {
			email: "",
			password: "",
			display_name: "",
			...birthPartsDefaults(),
		},
	});

	if (challenge) {
		return (
			<AuthEmailOtpStep
				challenge={challenge}
				leadKey="register.code_lead"
				submitKey="register.confirm_code"
				submittingKey="register.confirming_code"
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
					register.mutate(toRegisterRequest(values, i18n.language), {
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
					name="email"
					type="email"
					label={t("login.email")}
					placeholder={t("register.email_placeholder")}
					autoComplete="email"
					required
					labelClassName={authFieldLabelClass}
					inputClassName={authInputClass}
				/>
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
				{register.error ? (
					<p className="text-destructive text-[11px] font-medium">
						{translateError(t, register.error)}
					</p>
				) : null}
				<Button
					type="submit"
					disabled={register.isPending}
					className={`mt-2 ${authSubmitClass}`}
				>
					{register.isPending ? t("register.submitting") : t("register.submit")}
				</Button>
			</form>
		</Form>
	);
}
