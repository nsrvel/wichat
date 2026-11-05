"use client";

import { createInstance, type i18n as I18nInstance } from "i18next";
import { useEffect, useState } from "react";
import { I18nextProvider, initReactI18next } from "react-i18next";
import { type AppLocale } from "./locale";
import en from "./locales/en.json";
import id from "./locales/id.json";

const resources = {
	en: { translation: en },
	id: { translation: id },
};

/** Bundled resources init synchronously — safe for SSR + first paint (no useEffect-only init). */
export function createI18n(locale: AppLocale): I18nInstance {
	const instance = createInstance();
	instance.use(initReactI18next).init({
		lng: locale,
		fallbackLng: "en",
		resources,
		interpolation: { escapeValue: false },
		react: { useSuspense: false },
	});
	return instance;
}

export function I18nProvider({
	locale,
	children,
}: {
	locale: AppLocale;
	children: React.ReactNode;
}) {
	const [instance] = useState(() => createI18n(locale));

	useEffect(() => {
		if (instance.resolvedLanguage !== locale && instance.language !== locale) {
			void instance.changeLanguage(locale);
		}
	}, [instance, locale]);

	return <I18nextProvider i18n={instance}>{children}</I18nextProvider>;
}
