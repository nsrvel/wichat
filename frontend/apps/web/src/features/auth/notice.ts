const noticeCode = /^(auth|error|user)\.[a-z0-9_.]+$/;

export function safeNotice(code: string | undefined) {
	if (!code || !noticeCode.test(code)) {
		return undefined;
	}

	return code;
}
