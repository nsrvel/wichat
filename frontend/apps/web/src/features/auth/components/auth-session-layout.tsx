"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { AuthShell } from "@/features/auth/components/auth-shell";
import { AuthFormSkeleton } from "@/features/auth/components/auth-form-skeleton";
import { useMounted } from "@/hooks/use-mounted";
import { translateError } from "@/integrations/i18n/translate-error";
import * as authQuery from "@/services/auth/auth.query";

type AuthSessionLayoutProps = {
	headline: string;
	loadingHeadline: string;
	lead?: ReactNode;
	inviteBanner?: ReactNode;
	notice?: string;
	redirectIfUnauthenticated?: boolean;
	children: ReactNode;
};

export function AuthSessionLayout({
	headline,
	loadingHeadline,
	lead,
	inviteBanner,
	notice,
	redirectIfUnauthenticated = false,
	children,
}: AuthSessionLayoutProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const mounted = useMounted();
	const session = useQuery({ ...authQuery.sessionQueryOpts(), enabled: mounted });

	useEffect(() => {
		if (session.data) {
			router.replace("/");
		}
	}, [router, session.data]);

	useEffect(() => {
		if (!redirectIfUnauthenticated || session.isPending || session.data) {
			return;
		}
		if (!session.isPending && !session.data) {
			router.replace("/sign-in");
		}
	}, [redirectIfUnauthenticated, router, session.data, session.isPending]);

	const authLoading = !mounted || session.isPending;

	if (session.data) {
		return <div className="flex-1" aria-hidden suppressHydrationWarning />;
	}

	if (redirectIfUnauthenticated && !session.isPending && !session.data) {
		return <div className="flex-1" aria-hidden suppressHydrationWarning />;
	}

	if (session.isError) {
		return (
			<p className="text-destructive p-6 text-sm">{translateError(t, session.error)}</p>
		);
	}

	return (
		<AuthShell
			headline={authLoading ? loadingHeadline : headline}
			lead={authLoading ? null : lead}
			inviteBanner={authLoading ? null : inviteBanner}
			notice={notice ? t(notice, { defaultValue: t("error.unknown") }) : null}
		>
			{authLoading ? <AuthFormSkeleton /> : children}
		</AuthShell>
	);
}
