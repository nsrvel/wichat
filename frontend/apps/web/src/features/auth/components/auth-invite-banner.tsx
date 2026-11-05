"use client";

import { useTranslation } from "react-i18next";

export function AuthInviteBanner({ serverName }: { serverName: string }) {
	const { t } = useTranslation();
	return (
		<div className="mb-6 w-full min-w-0 rounded-xl border border-border bg-card/90 p-3.5 text-left text-sm sm:rounded-2xl sm:p-4">
			<span className="text-muted-foreground">{t("auth.invite_banner_label")}</span>
			<div className="mt-1 font-semibold text-foreground">{serverName}</div>
		</div>
	);
}
