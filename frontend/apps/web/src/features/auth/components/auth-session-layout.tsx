"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { AuthShell } from "@/features/auth/components/auth-shell";
import { authPageBackdropClass } from "@/lib/form-field-styles";
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
	authenticatedRedirectTo?: string;
	redirectIfUnauthenticated?: boolean;
	children: ReactNode;
};

export function AuthSessionLayout({
	headline,
	loadingHeadline,
	lead,
	inviteBanner,
	notice,
	authenticatedRedirectTo = "/",
	redirectIfUnauthenticated = false,
	children,
}: AuthSessionLayoutProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const mounted = useMounted();
	const session = useQuery({ ...authQuery.sessionQueryOpts(), enabled: mounted });

	useEffect(() => {
		if (session.data) {
			router.replace(authenticatedRedirectTo);
		}
	}, [authenticatedRedirectTo, router, session.data]);

	useEffect(() => {
		if (!redirectIfUnauthenticated || session.isPending || session.data) {
			return;
		}
		if (!session.isPending && !session.data) {
			router.replace("/sign-in");
		}
	}, [redirectIfUnauthenticated, router, session.data, session.isPending]);

	if (mounted && session.data) {
		return <div className={authPageBackdropClass} aria-hidden suppressHydrationWarning />;
	}

	if (redirectIfUnauthenticated && !session.isPending && !session.data) {
		return <div className={authPageBackdropClass} aria-hidden suppressHydrationWarning />;
	}

	if (session.isError && redirectIfUnauthenticated) {
		return <p className="text-destructive p-6 text-sm">{translateError(t, session.error)}</p>;
	}

	// Session refetch must not swap the whole card — skeleton only until client mount (SSR/hydration).
	const shellReady = mounted;

	return (
		<AuthShell
			headline={shellReady ? headline : loadingHeadline}
			lead={shellReady ? lead : null}
			inviteBanner={shellReady ? inviteBanner : null}
			notice={notice ? t(notice, { defaultValue: t("error.unknown") }) : null}
		>
			{shellReady ? children : <AuthFormSkeleton />}
		</AuthShell>
	);
}
