import { readStoredTimeFormat, resolveHour12, type TimeFormatPreference } from "@/lib/time-format";

const timeFormatterCache = new Map<string, Intl.DateTimeFormat>();

function getTimeFormatter(preference: TimeFormatPreference): Intl.DateTimeFormat {
	const hour12 = resolveHour12(preference);
	const key = hour12 === undefined ? "auto" : hour12 ? "12" : "24";
	let formatter = timeFormatterCache.get(key);
	if (!formatter) {
		formatter = new Intl.DateTimeFormat(undefined, {
			hour: "numeric",
			minute: "2-digit",
			...(hour12 === undefined ? {} : { hour12 }),
		});
		timeFormatterCache.set(key, formatter);
	}
	return formatter;
}

export function formatMessageTimeLabel(
	iso: string,
	preference: TimeFormatPreference = readStoredTimeFormat(),
): string {
	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) {
		return "";
	}
	return getTimeFormatter(preference).format(date);
}
