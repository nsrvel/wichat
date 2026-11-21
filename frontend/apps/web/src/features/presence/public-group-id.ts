const PUBLIC_GROUP_ID = /^[0-9a-f]{24}$/i;

export function normalizePublicGroupId(raw?: string | null): string | undefined {
	const id = raw?.trim().toLowerCase();
	return id && PUBLIC_GROUP_ID.test(id) ? id : undefined;
}
