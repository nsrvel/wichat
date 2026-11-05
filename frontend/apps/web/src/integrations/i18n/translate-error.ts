import { ApiError } from "@wichat/api-client/api-error";
import { isClientErrorCode } from "@wichat/api-client/client-error-code";
import type { TFunction } from "i18next";

export function translateError(t: TFunction, error: unknown): string {
	const fallback = t("error.unknown");

	if (error instanceof ApiError) {
		const code = isClientErrorCode(error.code) ? error.code : "error.unknown";

		if (code === "error.invalid_input") {
			return t("error.invalid_input", { defaultValue: fallback });
		}

		if (code === "auth.reset_rate_limited") {
			return t(code, {
				...(error.params ?? {}),
				defaultValue: fallback,
			});
		}

		return t(code, {
			...(error.params ?? {}),
			defaultValue: fallback,
		});
	}

	return fallback;
}
