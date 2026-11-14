import { ApiError } from "@wichat/api-client/api-error";
import type { TFunction } from "i18next";
import type { FieldPath, FieldValues, UseFormReturn } from "react-hook-form";
import { toast } from "sonner";
import { translateError } from "@/integrations/i18n/translate-error";

type ApplySettingsFormErrorOptions<T extends FieldValues> = {
	passwordField?: FieldPath<T>;
	codeMap?: Partial<Record<string, FieldPath<T>>>;
	paramFieldMap?: Partial<Record<string, FieldPath<T>>>;
};

/**
 * Maps API errors onto matching react-hook-form fields so validation/password errors
 * render inline via `<FormMessage />` instead of falling back to a toast.
 */
export function applySettingsFormError<T extends FieldValues>(
	form: UseFormReturn<T>,
	err: unknown,
	t: TFunction,
	options: ApplySettingsFormErrorOptions<T> = {},
): boolean {
	if (err instanceof ApiError) {
		if (err.code === "auth.invalid_credentials" && options.passwordField) {
			form.setError(options.passwordField, {
				type: "server",
				message: t("user_settings.account.incorrect_current_password"),
			});
			return true;
		}

		const mappedByCode = options.codeMap?.[err.code];
		if (mappedByCode) {
			form.setError(mappedByCode, {
				type: "server",
				message: translateError(t, err),
			});
			return true;
		}

		const paramField =
			typeof err.params?.field === "string" ? options.paramFieldMap?.[err.params.field] : undefined;
		if (paramField) {
			form.setError(paramField, {
				type: "server",
				message: translateError(t, err),
			});
			return true;
		}
	}

	toast.error(translateError(t, err));
	return false;
}
