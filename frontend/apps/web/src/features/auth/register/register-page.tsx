"use client";

import Link from "next/link";
import { useTranslation } from "react-i18next";
import { AuthInviteBanner } from "@/features/auth/components/auth-invite-banner";
import { AuthPrivacyNotice } from "@/features/auth/components/auth-privacy-notice";
import {
	AuthOr,
	authLeadSwitchClass,
	authLinkClass,
} from "@/features/auth/components/auth-shell";
import { AuthSessionLayout } from "@/features/auth/components/auth-session-layout";
import { GoogleButton } from "@/features/auth/components/google-button";
import { RegisterForm } from "./register-form";
import { authPath } from "@/lib/auth-invite";

export function RegisterPage({
	notice,
	inviteToken,
	serverName,
}: {
	notice?: string;
	inviteToken?: string;
	serverName?: string;
}) {
	const { t } = useTranslation();
	const isInviteFlow = Boolean(inviteToken || serverName);
	const loginHref = isInviteFlow
		? authPath("/sign-in", { invite: inviteToken, as: "account" })
		: "/sign-in";

	const lead = (
		<p className={authLeadSwitchClass}>
			<span>{t("auth.have_account")}</span>
			<Link href={loginHref} className={authLinkClass}>
				{t("auth.to_login")}
			</Link>
		</p>
	);

	const inviteBanner =
		serverName && isInviteFlow ? <AuthInviteBanner serverName={serverName} /> : null;

	return (
		<AuthSessionLayout
			headline={t("auth.title_register")}
			loadingHeadline={t("auth.title_register")}
			lead={lead}
			inviteBanner={inviteBanner}
			notice={notice}
		>
			<GoogleButton label={t("auth.google")} />
			<AuthOr label={t("auth.or")} />
			<RegisterForm />
			<AuthPrivacyNotice />
		</AuthSessionLayout>
	);
}
