"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { authLeadSwitchClass, authLinkClass } from "@/lib/form-field-styles";
import { AuthSessionLayout } from "@/features/auth/components/auth-session-layout";
import { RegisterForm } from "./register-form";
import { authPathWithReturn, postAuthRedirectPath } from "@/lib/auth-invite";
import {
	inviteTokenFromReturnPath,
	isInviteOnlyRegistrationBlocked,
} from "@/features/auth/lib/registration-policy";
import { instancePolicyQueryOpts } from "@/services/auth/auth.query";
import { StatusScreen } from "@/components/status/status-screen";
import { AuthFormSkeleton } from "@/features/auth/components/auth-form-skeleton";

export function RegisterPage({
	notice,
	authReturnTo = "/",
}: {
	notice?: string;
	authReturnTo?: string;
}) {
	const { t } = useTranslation();
	const fromInvite = authReturnTo.startsWith("/invite/");
	const inviteToken = inviteTokenFromReturnPath(authReturnTo);
	const { data: policy, isSuccess: policyReady } = useQuery(instancePolicyQueryOpts());

	const loginHref = fromInvite ? authPathWithReturn("/sign-in", authReturnTo) : "/sign-in";

	const lead = (
		<p className={authLeadSwitchClass}>
			<span>{t("auth.have_account")}</span>
			<Link href={loginHref} className={authLinkClass}>
				{t("auth.to_login")}
			</Link>
		</p>
	);

	const blocked = policyReady && isInviteOnlyRegistrationBlocked(policy, inviteToken);

	if (!policyReady) {
		return (
			<AuthSessionLayout
				headline={t("auth.title_register")}
				loadingHeadline={t("auth.title_register")}
				notice={notice}
				authenticatedRedirectTo={postAuthRedirectPath(authReturnTo)}
			>
				<AuthFormSkeleton />
			</AuthSessionLayout>
		);
	}

	if (blocked) {
		return <StatusScreen variant="forbidden" hideActions />;
	}

	return (
		<AuthSessionLayout
			headline={t("auth.title_register")}
			loadingHeadline={t("auth.title_register")}
			lead={lead}
			notice={notice}
			authenticatedRedirectTo={postAuthRedirectPath(authReturnTo)}
		>
			<RegisterForm authReturnTo={authReturnTo} inviteToken={inviteToken} />
		</AuthSessionLayout>
	);
}
