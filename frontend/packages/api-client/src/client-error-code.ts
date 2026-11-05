/** Matches i18n keys the API may return (auth / error / user). */
const clientErrorCode = /^(auth|error|user)\.[a-z0-9_.]+$/;

export function isClientErrorCode(code: string): boolean {
	return clientErrorCode.test(code);
}
