import { z } from "zod";

/** Matches backend `maxDisplayNameRunes` / `utf8.RuneCountInString`. */
export const MAX_DISPLAY_NAME_RUNES = 80;

export function unicodeRuneCount(value: string): number {
	let count = 0;
	for (const _ of value) {
		count++;
	}
	return count;
}

export const displayNameSchema = z
	.string()
	.trim()
	.min(1, "validation.display_name")
	.refine(
		(v) => unicodeRuneCount(v) <= MAX_DISPLAY_NAME_RUNES,
		"validation.display_name_length",
	);
