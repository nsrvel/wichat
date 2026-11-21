"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { authLeadSwitchClass, authLinkClass } from "@/lib/form-field-styles";
import { AuthSessionLayout } from "@/features/auth/components/auth-session-layout";
import { LoginForm } from "./login-form";
import { authPathWithReturn, postAuthRedirectPath } from "@/lib/auth-invite";
import { instancePolicyQueryOpts } from "@/services/auth/auth.query";

export function LoginPage({
	notice,
	authReturnTo = "/",
}: {
	notice?: string;
	authReturnTo?: string;
}) {
	const { t } = useTranslation();
	const fromInvite = authReturnTo.startsWith("/invite/");
	const { data: policy, isSuccess: policyReady } = useQuery(instancePolicyQueryOpts());

	const registerHref = fromInvite ? authPathWithReturn("/sign-up", authReturnTo) : "/sign-up";

	const lead =
		policyReady && policy?.invite_only ? null : (
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
			<LoginForm authReturnTo={authReturnTo} />
		</AuthSessionLayout>
	);
}
