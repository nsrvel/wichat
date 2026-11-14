"use client";

import Link from "next/link";
import { useTranslation } from "react-i18next";
import { AuthPrivacyNotice } from "@/features/auth/components/auth-privacy-notice";
import { AuthOr } from "@/features/auth/components/auth-shell";
import { authLeadSwitchClass, authLinkClass } from "@/lib/form-field-styles";
import { AuthSessionLayout } from "@/features/auth/components/auth-session-layout";
import { GoogleButton } from "@/features/auth/components/google-button";
import { LoginForm } from "./login-form";
import { authPathWithReturn, postAuthRedirectPath } from "@/lib/auth-invite";

export function LoginPage({
	notice,
	authReturnTo = "/",
}: {
	notice?: string;
	authReturnTo?: string;
}) {
	const { t } = useTranslation();
	const fromInvite = authReturnTo.startsWith("/invite/");

	const registerHref = fromInvite ? authPathWithReturn("/sign-up", authReturnTo) : "/sign-up";

	const lead = (
		<p className={authLeadSwitchClass}>
			<span>{t("auth.no_account")}</span>
			<Link href={registerHref} className={authLinkClass}>
				{t("auth.register")}
			</Link>
		</p>
	);

	return (
		<AuthSessionLayout
			headline={t("auth.title_login")}
			loadingHeadline={t("auth.title_login")}
			lead={lead}
			notice={notice}
			authenticatedRedirectTo={postAuthRedirectPath(authReturnTo)}
		>
			<GoogleButton label={t("auth.google")} authReturnTo={authReturnTo} />
			<AuthOr label={t("auth.or")} />
			<LoginForm />
			<AuthPrivacyNotice />
		</AuthSessionLayout>
	);
}
