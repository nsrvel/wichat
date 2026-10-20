"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { displayNameSchema, type GuestRegisterRequest } from "@wichat/contracts";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { BirthDateField } from "@/components/form/birth-date-field";
import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import { AuthPrivacyNotice } from "@/features/auth/components/auth-privacy-notice";
import {
	authFieldLabelClass,
	authFormClass,
	authInputClass,
	authSubmitClass,
} from "@/features/auth/components/auth-shell";
import { translateError } from "@/integrations/i18n/translate-error";
import { birthPartsDefaults, formatBirthDate, refineMinAgeBirth } from "@/lib/auth-validators";
import { useRegisterGuestMutation } from "@/services/auth/auth.mutation";

const birthDaySchema = z.string().regex(/^\d{2}$/);
const birthMonthSchema = z.string().regex(/^\d{2}$/);
const birthYearSchema = z.string().regex(/^\d{4}$/);

const GuestRegisterFormSchema = z
	.object({
		display_name: displayNameSchema,
		birth_day: birthDaySchema,
		birth_month: birthMonthSchema,
		birth_year: birthYearSchema,
	})
	.superRefine(refineMinAgeBirth);

type GuestRegisterFormValues = z.infer<typeof GuestRegisterFormSchema>;

function toGuestRegisterRequest(values: GuestRegisterFormValues): GuestRegisterRequest {
	return {
		display_name: values.display_name,
		birth_date: formatBirthDate(values.birth_year, values.birth_month, values.birth_day),
	};
}

export function GuestForm() {
	const { t } = useTranslation();
	const router = useRouter();
	const registerGuest = useRegisterGuestMutation();
	const form = useForm<GuestRegisterFormValues>({
		resolver: zodResolver(GuestRegisterFormSchema),
		defaultValues: {
			display_name: "",
			...birthPartsDefaults(),
		},
	});

	return (
		<Form {...form}>
			<form
				className={authFormClass}
				onSubmit={form.handleSubmit((values) => {
					registerGuest.mutate(toGuestRegisterRequest(values), {
						onSuccess: () => router.push("/"),
					});
				})}
			>
				<TextField
					control={form.control}
					name="display_name"
					label={t("guest.display_name")}
					placeholder={t("register.display_placeholder")}
					autoComplete="nickname"
					required
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
				{registerGuest.error ? (
					<p className="text-destructive text-[11px] font-medium">
						{translateError(t, registerGuest.error)}
					</p>
				) : null}
				<Button
					type="submit"
					disabled={registerGuest.isPending}
					className={`mt-2 ${authSubmitClass}`}
				>
					{registerGuest.isPending ? t("guest.submitting") : t("guest.submit")}
				</Button>
				<AuthPrivacyNotice />
			</form>
		</Form>
	);
}
