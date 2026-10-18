"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useMounted } from "@/features/auth/use-mounted";
import { translateError } from "@/integrations/i18n/translate-error";
import { authQuery } from "@/services/auth";
import { useTranslation } from "react-i18next";

export function AuthGate({ children }: { children: React.ReactNode }) {
	const { t } = useTranslation();
	const router = useRouter();
	const mounted = useMounted();
	const status = useQuery({ ...authQuery.setupStatusQueryOpts(), enabled: mounted });
	const session = useQuery({
		...authQuery.sessionQueryOpts(),
		enabled: mounted && status.data?.needs_setup === false,
	});

	useEffect(() => {
		if (!status.data) {
			return;
		}
		if (status.data.needs_setup) {
			router.replace("/sign-up");
			return;
		}
		if (session.isError) {
			router.replace("/sign-in");
		}
	}, [router, session.isError, status.data]);

	if (status.isError) {
		return <p className="text-destructive p-6 text-sm">{translateError(t, status.error)}</p>;
	}
	if (!mounted || status.isPending || (status.data?.needs_setup === false && session.isPending)) {
		return <div className="flex-1" />;
	}
	if (!session.data) {
		return <div className="flex-1" />;
	}

	return children;
}
