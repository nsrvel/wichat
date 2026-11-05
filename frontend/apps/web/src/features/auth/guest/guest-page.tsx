"use client";

import Link from "next/link";
import { useTranslation } from "react-i18next";
import { AuthInviteBanner } from "@/features/auth/components/auth-invite-banner";
import { authLeadSwitchClass, authLinkClass } from "@/features/auth/components/auth-shell";
import { AuthSessionLayout } from "@/features/auth/components/auth-session-layout";
import { GuestForm } from "./guest-form";
import { authPath } from "@/lib/auth-invite";

export function GuestPage({
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
	const loginHref = authPath("/sign-in", { invite: inviteToken, as: "account" });

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
			headline={t("auth.title_guest")}
			loadingHeadline={t("auth.title_guest")}
			lead={lead}
			inviteBanner={inviteBanner}
			notice={notice}
			redirectIfUnauthenticated={!isInviteFlow}
		>
			<GuestForm />
		</AuthSessionLayout>
	);
}
