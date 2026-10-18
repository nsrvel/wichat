"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
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
import { LoginRequestSchema, type LoginRequest } from "@wichat/contracts";
import { authQuery } from "@/services/auth";
import { useLoginMutation } from "@/services/auth/auth.mutation";

export function LoginPage({ notice }: { notice?: string }) {
	const { t } = useTranslation();
	const router = useRouter();
	const mounted = useMounted();
	const login = useLoginMutation();
	const status = useQuery({ ...authQuery.setupStatusQueryOpts(), enabled: mounted });
	const session = useQuery({
		...authQuery.sessionQueryOpts(),
		enabled: mounted && status.data?.needs_setup === false,
	});
	const form = useForm<LoginRequest>({
		resolver: zodResolver(LoginRequestSchema),
		defaultValues: { email: "", password: "" },
	});

	useEffect(() => {
		if (session.data) {
			router.replace("/");
		}
	}, [router, session.data]);

	const sessionEnabled = mounted && status.data?.needs_setup === false;
	if (!mounted || status.isPending || (sessionEnabled && session.isPending) || session.data) {
		return <div className="flex-1" />;
	}
	if (status.isError) {
		return <p className="text-destructive p-6 text-sm">{translateError(t, status.error)}</p>;
	}

	return (
		<AuthShell
			headline={t("login.title")}
			notice={notice ? t(notice, { defaultValue: t("error.unknown") }) : null}
			lead={
				status.data?.needs_setup ? (
					<p>
						{t("login.first_time")}{" "}
						<Link href="/sign-up" className={authLinkClass}>
							{t("login.sign_up")}
						</Link>
					</p>
				) : null
			}
		>
			<form
				className="flex flex-col gap-3"
				onSubmit={form.handleSubmit((values) => {
					login.mutate(values, { onSuccess: () => router.push("/") });
				})}
			>
				<div className="flex flex-col gap-1.5">
					<Label htmlFor="email" className="sr-only">
						{t("login.email")}
					</Label>
					<Input
						id="email"
						type="email"
						autoComplete="email"
						placeholder={t("login.email_placeholder")}
						className={authInputClass}
						{...form.register("email")}
					/>
					{form.formState.errors.email ? (
						<p className="text-destructive text-[11px] font-medium">{t("validation.email")}</p>
					) : null}
				</div>
				<PasswordField
					id="password"
					label={t("login.password")}
					placeholder={t("login.password_placeholder")}
					autoComplete="current-password"
					registration={form.register("password")}
					error={form.formState.errors.password ? t("validation.password") : undefined}
					showLabel={t("login.show_password")}
					hideLabel={t("login.hide_password")}
				/>
				{login.error ? (
					<p className="text-destructive text-[11px] font-medium">
						{translateError(t, login.error)}
					</p>
				) : null}
				<Button type="submit" disabled={login.isPending} className={`mt-1 ${authSubmitClass}`}>
					{login.isPending ? t("login.submitting") : t("login.submit")}
				</Button>
			</form>
			<AuthOr label={t("auth.or")} />
			<GoogleButton label={t("auth.google")} />
		</AuthShell>
	);
}
