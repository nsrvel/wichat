"use client";

import { isAppLocale, SUPPORTED_LOCALES, type AppLocale } from "@wichat/i18n/locale";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { CustomSelect, type SelectOption } from "@/components/ui/custom-select";
import { applyAppLocale } from "@/integrations/i18n/apply-app-locale";
import { translateError } from "@/integrations/i18n/translate-error";
import {
	useTimeFormatPreference,
	writeStoredTimeFormat,
	type TimeFormatPreference,
} from "@/lib/time-format";
import { cn } from "@/lib/utils";
import { useUpdateMyPreferencesMutation } from "@/services/user/user.mutation";

const LOCALE_FLAGS: Record<AppLocale, string> = {
	en: "🇺🇸",
	id: "🇮🇩",
};

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
	const activeLocale = resolveActiveLocale(i18n.resolvedLanguage ?? i18n.language);
	const timeFormat = useTimeFormatPreference();
	const updatePreferences = useUpdateMyPreferencesMutation();

	const languageOptions: SelectOption[] = useMemo(
		() =>
			SUPPORTED_LOCALES.map((locale) => ({
				value: locale.code,
				label: `${LOCALE_FLAGS[locale.code]}  ${locale.nativeLabel}`,
			})),
		[],
	);

	const timeFormatOptions: { value: TimeFormatPreference; label: string }[] = [
		{ value: "auto", label: t("user_settings.language.time_format_auto") },
		{ value: "12h", label: t("user_settings.language.time_format_12h") },
		{ value: "24h", label: t("user_settings.language.time_format_24h") },
	];

	const handleLocaleChange = async (nextLocale: AppLocale) => {
		if (nextLocale === activeLocale) return;
		const previousLocale = activeLocale;
		await applyAppLocale(nextLocale, i18n);
		try {
			await updatePreferences.mutateAsync({
				locale: nextLocale,
				time_format: timeFormat,
			});
		} catch (err) {
			await applyAppLocale(previousLocale, i18n);
			toast.error(translateError(t, err));
		}
	};

	const handleTimeFormatChange = async (nextFormat: TimeFormatPreference) => {
		if (nextFormat === timeFormat) return;
		const previousFormat = timeFormat;
		writeStoredTimeFormat(nextFormat);
		try {
			await updatePreferences.mutateAsync({
				locale: activeLocale,
				time_format: nextFormat,
			});
		} catch (err) {
			writeStoredTimeFormat(previousFormat);
			toast.error(translateError(t, err));
		}
	};

	return (
		<div className="space-y-7">
			<div id="user-settings-language-select" className="space-y-3">
				<h3 className="text-sm font-semibold text-foreground">
					{t("user_settings.language.select_label")}
				</h3>

				<CustomSelect
					options={languageOptions}
					value={activeLocale}
					placeholder={t("user_settings.language.select_label")}
					onChange={(val) => {
						if (isAppLocale(val)) {
							void handleLocaleChange(val);
						}
					}}
					portaled
					triggerClassName="flex h-11 w-full items-center justify-between rounded-xl border border-input bg-field px-3.5 text-sm font-medium text-foreground transition-colors hover:border-foreground/30 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40"
					menuClassName="z-[90]"
				/>
			</div>

			<div id="user-settings-time-format" className="space-y-3">
				<h3 className="text-sm font-semibold text-foreground">
					{t("user_settings.language.time_format_label")}
				</h3>

				<div
					role="radiogroup"
					aria-label={t("user_settings.language.time_format_label")}
					className="flex flex-col gap-2.5"
				>
					{timeFormatOptions.map((option) => {
						const selected = timeFormat === option.value;
						return (
							<button
								key={option.value}
								type="button"
								role="radio"
								aria-checked={selected}
								onClick={() => void handleTimeFormatChange(option.value)}
								className="group flex w-fit items-center gap-3 py-0.5 text-left focus-visible:outline-none"
							>
								<span
									className={cn(
										"flex size-5 shrink-0 items-center justify-center rounded-full border-2 transition-colors",
										selected
											? "border-primary bg-primary"
											: "border-muted-foreground/50 bg-transparent group-hover:border-foreground/70",
									)}
									aria-hidden
								>
									{selected ? <span className="size-2 rounded-full bg-primary-foreground" /> : null}
								</span>
								<span
									className={cn(
										"text-sm font-medium transition-colors",
										selected
											? "text-foreground"
											: "text-muted-foreground group-hover:text-foreground",
									)}
								>
									{option.label}
								</span>
							</button>
						);
					})}
				</div>
			</div>
		</div>
	);
}
