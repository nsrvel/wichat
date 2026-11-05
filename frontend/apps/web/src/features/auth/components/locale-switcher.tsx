"use client";

import { isAppLocale, SUPPORTED_LOCALES, type AppLocale } from "@wichat/i18n/locale";
import { useCallback, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { CustomSelect } from "@/components/ui/custom-select";
import { authSelectGhostTriggerClass } from "@/features/auth/components/auth-shell";
import { applyAppLocale } from "@/integrations/i18n/apply-app-locale";
import { appSans } from "@/lib/fonts/app-sans";
import { cn } from "@/lib/utils";

const localeFont = appSans.className;

function activeLocaleFromI18n(language: string | undefined): AppLocale {
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
	const prefix = SUPPORTED_LOCALES.find((entry) => language.startsWith(`${entry.code}-`));
	return prefix?.code ?? "en";
}

function useLocaleSwitch() {
	const { t, i18n } = useTranslation();
	const active = activeLocaleFromI18n(i18n.resolvedLanguage ?? i18n.language);

	const choose = useCallback(
		async (locale: AppLocale) => {
			if (locale === active) {
				return;
			}
			await applyAppLocale(locale, i18n);
		},
		[active, i18n],
	);

	return { t, active, choose };
}

export function LocaleSwitcher() {
	const { t, active, choose } = useLocaleSwitch();

	const options = useMemo(
		() =>
			SUPPORTED_LOCALES.map((entry) => ({
				value: entry.code,
				label: entry.nativeLabel,
				triggerLabel: entry.code.toUpperCase(),
			})),
		[],
	);

	return (
		<CustomSelect
			className={cn(localeFont, "w-auto shrink-0")}
			menuClassName={localeFont}
			options={options}
			value={active}
			placeholder={t("locale.label")}
			onChange={(value) => void choose(value as AppLocale)}
			triggerClassName={cn(localeFont, authSelectGhostTriggerClass)}
			compactTrigger
			portaled
			menuAlign="end"
		/>
	);
}
