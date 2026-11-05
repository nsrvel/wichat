import type { AppLocale } from "@wichat/i18n/locale";
import type { i18n as I18nInstance } from "i18next";
import { writeLocaleCookie } from "@/integrations/i18n/locale";

/** Client locale switch: cookie + i18n + document lang (no router.refresh — avoids RSC fighting client state). */
export async function applyAppLocale(locale: AppLocale, i18n: I18nInstance) {
	writeLocaleCookie(locale);
	await i18n.changeLanguage(locale);
	document.documentElement.lang = locale;
}
