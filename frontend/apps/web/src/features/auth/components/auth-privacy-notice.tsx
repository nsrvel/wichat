"use client";

import Link from "next/link";
import { useTranslation } from "react-i18next";

export function AuthPrivacyNotice() {
	const { t } = useTranslation();

	return (
		<p className="px-1 pt-1 pb-0.5 text-center text-[11px] leading-snug text-subtle-foreground sm:text-xs max-h-[880px]:pt-0.5 max-h-[880px]:leading-tight">
			{t("auth.privacy_prefix")}{" "}
			<Link
				href="/privacy"
				target="_blank"
				rel="noopener noreferrer"
				className="text-foreground/90 underline decoration-subtle-foreground/40 underline-offset-2 hover:decoration-foreground"
			>
				{t("auth.privacy_link")}
			</Link>
			{t("auth.privacy_suffix")}
		</p>
	);
}
