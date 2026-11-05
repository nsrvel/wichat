import type { z } from "zod";

const currentYear = new Date().getFullYear();

export function birthPartsDefaults() {
	return {
		birth_day: "01",
		birth_month: "01",
		birth_year: String(currentYear - 18),
	};
}

export function refineMinAgeBirth(
	data: { birth_day: string; birth_month: string; birth_year: string },
	ctx: z.RefinementCtx,
) {
	if (!isAtLeast13(data.birth_year, data.birth_month, data.birth_day)) {
		ctx.addIssue({
			code: "custom",
			message: "auth.age_restricted",
			path: ["birth_year"],
		});
	}
}

export function formatBirthDate(year: string, month: string, day: string) {
	return `${year}-${month}-${day}`;
}

export function calculateAge(year: string, month: string, day: string) {
	const y = Number.parseInt(year, 10);
	const m = Number.parseInt(month, 10);
	const d = Number.parseInt(day, 10);
	const today = new Date();
	let age = today.getFullYear() - y;
	const hadBirthday =
		today.getMonth() + 1 > m || (today.getMonth() + 1 === m && today.getDate() >= d);
	if (!hadBirthday) {
		age--;
	}
	return age;
}

export function isAtLeast13(year: string, month: string, day: string) {
	return calculateAge(year, month, day) >= 13;
}
