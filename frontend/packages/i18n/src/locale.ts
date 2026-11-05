export const SUPPORTED_LOCALES = [
	{ code: "en", nativeLabel: "English" },
	{ code: "id", nativeLabel: "Indonesia" },
] as const;

export type AppLocale = (typeof SUPPORTED_LOCALES)[number]["code"];

export const DEFAULT_APP_LOCALE: AppLocale = "en";

const localeCodes = new Set<string>(
	SUPPORTED_LOCALES.map((entry) => entry.code),
);

export function isAppLocale(value: string): value is AppLocale {
	return localeCodes.has(value);
}

export function parseAppLocale(value: string | undefined): AppLocale {
	if (value && isAppLocale(value)) {
		return value;
	}
	return DEFAULT_APP_LOCALE;
}

export function appLocaleMeta(code: AppLocale) {
	return (
		SUPPORTED_LOCALES.find((entry) => entry.code === code) ??
		SUPPORTED_LOCALES[0]
	);
}
