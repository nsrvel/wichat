import { ApiError } from "@wichat/api-client";
import type { TFunction } from "i18next";

export function translateError(t: TFunction, error: unknown): string {
	if (error instanceof ApiError) {
		return t(error.code, {
			...(error.params ?? {}),
			defaultValue: t("error.unknown"),
		});
	}

	return t("error.unknown");
}
