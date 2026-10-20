"use client";

import { useTranslation } from "react-i18next";

export function AuthInviteBanner({ serverName }: { serverName: string }) {
	const { t } = useTranslation();
	return (
		<div className="mb-6 w-full min-w-0 rounded-xl border border-black/10 bg-white/90 p-3.5 text-left text-sm sm:rounded-2xl sm:p-4 dark:border-white/10 dark:bg-[#1a1a1a]">
			<span className="text-[#6d6d6d] dark:text-[#a3a3a3]">{t("auth.invite_banner_label")}</span>
			<div className="mt-1 font-semibold text-[#171717] dark:text-white">{serverName}</div>
		</div>
	);
}
