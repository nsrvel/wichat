/** Error UI only when the query failed and has no cached page to render. */
export function isQueryFailureWithoutData(isError: boolean, data: unknown): boolean {
	return isError && data === undefined;
}
