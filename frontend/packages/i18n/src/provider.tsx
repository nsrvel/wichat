"use client";

import { createInstance, type i18n as I18nInstance } from "i18next";
import { useState } from "react";
import { I18nextProvider, initReactI18next } from "react-i18next";
import { type AppLocale } from "./locale";
import en from "./locales/en.json";
import id from "./locales/id.json";

export function createI18n(locale: AppLocale): I18nInstance {
	const instance = createInstance();
	void instance.use(initReactI18next).init({
		lng: locale,
		fallbackLng: "en",
		resources: {
			en: { translation: en },
			id: { translation: id },
		},
		interpolation: { escapeValue: false },
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

	return <I18nextProvider i18n={instance}>{children}</I18nextProvider>;
}
