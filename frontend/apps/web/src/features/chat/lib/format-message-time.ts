import { readStoredTimeFormat, resolveHour12, type TimeFormatPreference } from "@/lib/time-format";

export function formatMessageTimeLabel(
	iso: string,
	preference: TimeFormatPreference = readStoredTimeFormat(),
): string {
	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) {
		return "";
	}
	return date.toLocaleTimeString(undefined, {
		hour: "numeric",
		minute: "2-digit",
		hour12: resolveHour12(preference),
	});
}
