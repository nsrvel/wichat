export function groupInitials(name: string): string {
	const parts = name.trim().split(/\s+/).filter(Boolean);
	if (parts.length >= 2) {
		return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
	}
	const trimmed = name.trim();
	return trimmed.slice(0, 2).toUpperCase() || "?";
}
