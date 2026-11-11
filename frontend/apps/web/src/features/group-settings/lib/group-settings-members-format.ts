export function parseUtcIso(iso: string): Date | null {
	if (!iso) {
		return null;
	}
	const d = new Date(iso);
	return Number.isNaN(d.getTime()) ? null : d;
}

export function formatShortDate(iso: string, locale: string): string {
	const d = parseUtcIso(iso);
	if (!d) {
		return "—";
	}
	return new Intl.DateTimeFormat(locale, {
		year: "numeric",
		month: "short",
		day: "numeric",
		timeZone: "UTC",
	}).format(d);
}

export function formatRelativeUtc(iso: string, locale: string, now = Date.now()): string {
	const d = parseUtcIso(iso);
	if (!d) {
		return "—";
	}
	const diffMs = d.getTime() - now;
	const diffSec = Math.round(diffMs / 1000);
	const absSec = Math.abs(diffSec);
	const rtf = new Intl.RelativeTimeFormat(locale, { numeric: "auto" });

	if (absSec < 60) {
		return rtf.format(diffSec, "second");
	}
	const diffMin = Math.round(diffSec / 60);
	if (Math.abs(diffMin) < 60) {
		return rtf.format(diffMin, "minute");
	}
	const diffHr = Math.round(diffSec / 3600);
	if (Math.abs(diffHr) < 48) {
		return rtf.format(diffHr, "hour");
	}
	const diffDay = Math.round(diffSec / 86400);
	if (Math.abs(diffDay) < 60) {
		return rtf.format(diffDay, "day");
	}
	const diffMonth = Math.round(diffSec / (86400 * 30));
	return rtf.format(diffMonth, "month");
}
