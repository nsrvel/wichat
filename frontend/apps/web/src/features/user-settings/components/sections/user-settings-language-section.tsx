"use client";

import { isAppLocale, SUPPORTED_LOCALES, type AppLocale } from "@wichat/i18n/locale";
import { Check } from "lucide-react";
import { useTranslation } from "react-i18next";
import { applyAppLocale } from "@/integrations/i18n/apply-app-locale";
import { cn } from "@/lib/utils";

function resolveActiveLocale(language: string | undefined): AppLocale {
	if (!language) {
		return "en";
	}
	const base = language.split("-")[0];
	if (isAppLocale(base)) {
		return base;
	}
	if (isAppLocale(language)) {
		return language;
	}
	return "en";
}

export function UserSettingsLanguageSection() {
	const { t, i18n } = useTranslation();
	const active = resolveActiveLocale(i18n.resolvedLanguage ?? i18n.language);

	return (
		<div className="space-y-5">
			<div id="user-settings-language-select" className="space-y-2">
				<p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
					{t("user_settings.language.select_label")}
				</p>
				<div className="divide-y divide-border/60 rounded-xl border border-border/70 bg-muted/20">
					{SUPPORTED_LOCALES.map((locale) => {
						const selected = locale.code === active;
						return (
							<button
								key={locale.code}
								type="button"
								onClick={() => void applyAppLocale(locale.code, i18n)}
								className={cn(
									"flex w-full items-center justify-between px-4 py-3 text-left transition-colors first:rounded-t-xl last:rounded-b-xl hover:bg-foreground/5",
									selected && "bg-foreground/5",
								)}
							>
								<div>
									<p className="text-sm font-medium text-foreground">{locale.nativeLabel}</p>
									<p className="text-xs text-muted-foreground uppercase">{locale.code}</p>
								</div>
								{selected ? <Check className="size-4 text-primary" aria-hidden /> : null}
							</button>
						);
					})}
				</div>
			</div>
		</div>
	);
}
