/** Month labels for the birth-date picker (follows app locale). */
export function birthDateMonthOptions(locale: string): { value: string; label: string }[] {
	const intlLocale = locale === "id" ? "id-ID" : "en-US";

	return Array.from({ length: 12 }, (_, i) => {
		const value = String(i + 1).padStart(2, "0");
		const label = new Intl.DateTimeFormat(intlLocale, { month: "long" }).format(
			new Date(2024, i, 1),
		);

		return { value, label };
	});
}
