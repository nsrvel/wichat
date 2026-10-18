"use client";

import { useRouter } from "next/navigation";
import { useTranslation } from "react-i18next";
import { writeLocaleCookie, type AppLocale } from "@/integrations/i18n/locale";

export function LocaleSwitcher() {
	const { t, i18n } = useTranslation();
	const router = useRouter();

	function choose(locale: AppLocale) {
		writeLocaleCookie(locale);
		void i18n.changeLanguage(locale);
		router.refresh();
	}

	return (
		<div className="flex items-center gap-2" role="group" aria-label={t("locale.label")}>
			{(["en", "id"] as const).map((locale) => {
				const selected = i18n.resolvedLanguage === locale || i18n.language.startsWith(locale);
				return (
					<button
						key={locale}
						type="button"
						aria-pressed={selected}
						onClick={() => choose(locale)}
						className={
							selected
								? "bg-primary text-primary-foreground rounded-lg px-2.5 py-1 text-xs font-medium"
								: "text-muted-foreground hover:bg-muted rounded-lg px-2.5 py-1 text-xs font-medium"
						}
					>
						{t(`locale.${locale}`)}
					</button>
				);
			})}
		</div>
	);
}
