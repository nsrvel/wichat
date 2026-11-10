const LEGACY_GENERAL = "general";

export const DISPLAY_NAME_ANNOUNCEMENTS = "Announcements";
export const DISPLAY_NAME_GENERAL = "General";

/** Until DB migration renames seeded rows; API may still return lowercase. */
export function formatChannelDisplayName(name: string): string {
	const lower = name.trim().toLowerCase();
	if (lower === "announcements") {
		return DISPLAY_NAME_ANNOUNCEMENTS;
	}
	if (lower === LEGACY_GENERAL) {
		return DISPLAY_NAME_GENERAL;
	}
	return name;
}
