"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useTranslation } from "react-i18next";
import { useMounted } from "@/hooks/use-mounted";
import * as authQuery from "@/services/auth/auth.query";
import { AuthInviteBanner } from "@/features/auth/components/auth-invite-banner";
import { AuthPrivacyNotice } from "@/features/auth/components/auth-privacy-notice";
import {
	AuthOr,
	authLeadSwitchClass,
	authLinkClass,
} from "@/features/auth/components/auth-shell";
import { AuthSessionLayout } from "@/features/auth/components/auth-session-layout";
import { GoogleButton } from "@/features/auth/components/google-button";
import { LoginForm } from "./login-form";
import { authPath } from "@/lib/auth-invite";

export function LoginPage({
	notice,
	inviteToken,
	serverName,
}: {
	notice?: string;
	inviteToken?: string;
	serverName?: string;
}) {
	const { t } = useTranslation();
	const mounted = useMounted();
	const policy = useQuery({ ...authQuery.instancePolicyQueryOpts(), enabled: mounted });
	const isInviteFlow = Boolean(inviteToken || serverName);
	const showGuest = policy.data?.guest_enabled === true;

	const lead =
		isInviteFlow && showGuest ? (
			<p className={authLeadSwitchClass}>
				<span>{t("auth.no_account")}</span>
				<Link href={authPath("/sign-in", { invite: inviteToken })} className={authLinkClass}>
					{t("guest.switcher")}
				</Link>
			</p>
		) : (
			<p className={authLeadSwitchClass}>
				<span>{t("auth.no_account")}</span>
				<Link href="/sign-up" className={authLinkClass}>
					{t("auth.register")}
				</Link>
			</p>
		);

	const inviteBanner =
		serverName && isInviteFlow ? <AuthInviteBanner serverName={serverName} /> : null;

	return (
		<AuthSessionLayout
			headline={t("auth.title_login")}
			loadingHeadline={t("auth.title_login")}
			lead={lead}
			inviteBanner={inviteBanner}
			notice={notice}
		>
			<GoogleButton label={t("auth.google")} />
			<AuthOr label={t("auth.or")} />
			<LoginForm />
			<AuthPrivacyNotice />
		</AuthSessionLayout>
	);
}
