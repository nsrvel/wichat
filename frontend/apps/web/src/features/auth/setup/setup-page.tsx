"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useForm, type UseFormRegisterReturn } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	AuthOr,
	AuthShell,
	authInputClass,
	authLinkClass,
	authSubmitClass,
} from "@/features/auth/components/auth-shell";
import { GoogleButton } from "@/features/auth/components/google-button";
import { PasswordField } from "@/features/auth/components/password-field";
import { useMounted } from "@/features/auth/use-mounted";
import { translateError } from "@/integrations/i18n/translate-error";
import { SetupRequestSchema, type SetupRequest } from "@wichat/contracts";
import { authQuery } from "@/services/auth";
import { useSetupMutation } from "@/services/auth/auth.mutation";

export function SetupPage({ notice }: { notice?: string }) {
	const { t } = useTranslation();
	const router = useRouter();
	const mounted = useMounted();
	const setup = useSetupMutation();
	const status = useQuery({ ...authQuery.setupStatusQueryOpts(), enabled: mounted });
	const form = useForm<SetupRequest>({
		resolver: zodResolver(SetupRequestSchema),
		defaultValues: { email: "", password: "", display_name: "" },
	});

	useEffect(() => {
		if (status.data && !status.data.needs_setup) {
			router.replace("/sign-in");
		}
	}, [router, status.data]);

	if (!mounted || status.isPending || (status.data && !status.data.needs_setup)) {
		return <div className="flex-1" />;
	}
	if (status.isError || !status.data) {
		return (
			<p className="text-destructive p-6 text-sm">
				{translateError(t, status.error ?? new Error("error.unknown"))}
			</p>
		);
	}

	return (
		<AuthShell
			headline={t("setup.title")}
			notice={notice ? t(notice, { defaultValue: t("error.unknown") }) : null}
			lead={
				<p>
					{t("auth.have_account")}{" "}
					<Link href="/sign-in" className={authLinkClass}>
						{t("auth.to_login")}
					</Link>
				</p>
			}
		>
			<form
				className="flex flex-col gap-3"
				onSubmit={form.handleSubmit((values) => {
					setup.mutate(values, { onSuccess: () => router.push("/") });
				})}
			>
				<Field
					id="email"
					label={t("login.email")}
					type="email"
					placeholder={t("login.email_placeholder")}
					autoComplete="email"
					error={form.formState.errors.email ? t("validation.email") : undefined}
					registration={form.register("email")}
				/>
				<Field
					id="display_name"
					label={t("setup.display_name")}
					placeholder={t("setup.display_placeholder")}
					autoComplete="name"
					error={form.formState.errors.display_name ? t("validation.display_name") : undefined}
					registration={form.register("display_name")}
				/>
				<PasswordField
					id="password"
					label={t("login.password")}
					placeholder={t("login.password_placeholder")}
					autoComplete="new-password"
					registration={form.register("password")}
					error={form.formState.errors.password ? t("validation.password") : undefined}
					showLabel={t("login.show_password")}
					hideLabel={t("login.hide_password")}
				/>
				{setup.error ? (
					<p className="text-destructive text-[11px] font-medium">
						{translateError(t, setup.error)}
					</p>
				) : null}
				<Button type="submit" disabled={setup.isPending} className={`mt-1 ${authSubmitClass}`}>
					{setup.isPending ? t("setup.submitting") : t("setup.submit")}
				</Button>
			</form>
			<AuthOr label={t("auth.or")} />
			<GoogleButton label={t("auth.google")} />
		</AuthShell>
	);
}

function Field({
	id,
	label,
	type = "text",
	placeholder,
	autoComplete,
	error,
	registration,
}: {
	id: string;
	label: string;
	type?: string;
	placeholder: string;
	autoComplete: string;
	error?: string;
	registration: UseFormRegisterReturn;
}) {
	return (
		<div className="flex flex-col gap-1.5">
			<Label htmlFor={id} className="sr-only">
				{label}
			</Label>
			<Input
				id={id}
				type={type}
				autoComplete={autoComplete}
				placeholder={placeholder}
				className={authInputClass}
				{...registration}
			/>
			{error ? <p className="text-destructive text-[11px] font-medium">{error}</p> : null}
		</div>
	);
}
