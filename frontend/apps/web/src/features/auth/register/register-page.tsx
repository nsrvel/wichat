"use client";

import Link from "next/link";
import { useTranslation } from "react-i18next";
import { AuthPrivacyNotice } from "@/features/auth/components/auth-privacy-notice";
import { AuthOr, authLeadSwitchClass, authLinkClass } from "@/features/auth/components/auth-shell";
import { AuthSessionLayout } from "@/features/auth/components/auth-session-layout";
import { GoogleButton } from "@/features/auth/components/google-button";
import { RegisterForm } from "./register-form";
import { authPathWithReturn, postAuthRedirectPath } from "@/lib/auth-invite";

export function RegisterPage({
	notice,
	authReturnTo = "/",
}: {
	notice?: string;
	authReturnTo?: string;
}) {
	const { t } = useTranslation();
	const fromInvite = authReturnTo.startsWith("/invite/");
	const loginHref = fromInvite ? authPathWithReturn("/sign-in", authReturnTo) : "/sign-in";

	const lead = (
		<p className={authLeadSwitchClass}>
			<span>{t("auth.have_account")}</span>
			<Link href={loginHref} className={authLinkClass}>
				{t("auth.to_login")}
			</Link>
		</p>
	);

	return (
		<AuthSessionLayout
			headline={t("auth.title_register")}
			loadingHeadline={t("auth.title_register")}
			lead={lead}
			notice={notice}
			authenticatedRedirectTo={postAuthRedirectPath(authReturnTo)}
		>
			<GoogleButton label={t("auth.google")} authReturnTo={authReturnTo} />
			<AuthOr label={t("auth.or")} />
			<RegisterForm authReturnTo={authReturnTo} />
			<AuthPrivacyNotice />
		</AuthSessionLayout>
	);
}
