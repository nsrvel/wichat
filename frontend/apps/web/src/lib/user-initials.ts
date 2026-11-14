export function userInitials(displayName: string, email?: string): string {
	const fromName = displayName.trim().split(/\s+/).filter(Boolean);
	if (fromName.length >= 2) {
		return `${fromName[0]![0] ?? ""}${fromName[1]![0] ?? ""}`.toUpperCase();
	}
	if (fromName.length === 1 && fromName[0]) {
		return fromName[0].slice(0, 2).toUpperCase();
	}
	const e = (email ?? "").trim();
	if (e) {
		return e.slice(0, 2).toUpperCase();
	}
	return "?";
}
