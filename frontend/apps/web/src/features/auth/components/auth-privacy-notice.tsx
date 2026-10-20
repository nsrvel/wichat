"use client";

import Link from "next/link";
import { useTranslation } from "react-i18next";

export function AuthPrivacyNotice() {
	const { t } = useTranslation();

	return (
		<p className="px-1 pt-1 pb-0.5 text-center text-[11px] leading-snug text-[#8a8a8a] sm:text-xs max-h-[880px]:pt-0.5 max-h-[880px]:leading-tight dark:text-[#a3a3a3]">
			{t("auth.privacy_prefix")}{" "}
			<Link
				href="/privacy"
				target="_blank"
				rel="noopener noreferrer"
				className="text-[#171717]/90 underline decoration-[#8a8a8a]/40 underline-offset-2 hover:decoration-[#171717] dark:text-white/90 dark:hover:decoration-white"
			>
				{t("auth.privacy_link")}
			</Link>
			{t("auth.privacy_suffix")}
		</p>
	);
}
