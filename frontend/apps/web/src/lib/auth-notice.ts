import { isClientErrorCode } from "@wichat/api-client/client-error-code";

export function safeNotice(code: string | undefined) {
	if (!code || !isClientErrorCode(code)) {
		return undefined;
	}

	return code;
}
