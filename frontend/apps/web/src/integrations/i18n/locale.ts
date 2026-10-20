import type { AppLocale } from "@wichat/i18n/locale";

export type { AppLocale } from "@wichat/i18n/locale";

export const LOCALE_COOKIE = "wichat_locale";

export function localeFromCookie(value: string | undefined): AppLocale {
	return value === "id" ? "id" : "en";
}

export function writeLocaleCookie(locale: AppLocale) {
	document.cookie = `${LOCALE_COOKIE}=${locale}; Path=/; Max-Age=31536000; SameSite=Lax`;
}
