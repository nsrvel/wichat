import type { TFunction } from "i18next";

const formI18nKey = /^(auth|error|user|validation)\.[a-z0-9_.]+$/;

/** Zod schemas store i18n keys in `message`; inline fields use FormMessage — footers must use this too. */
export function translateFormFieldMessage(
	t: TFunction,
	raw: string | undefined,
): string | undefined {
	if (!raw) {
		return undefined;
	}
	if (formI18nKey.test(raw)) {
		return t(raw, { defaultValue: t("error.unknown") });
	}
	return raw;
}
