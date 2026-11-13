import { z } from "zod";

/** Matches backend `MinDisplayNameRunes` / `MaxDisplayNameRunes`. */
export const MIN_DISPLAY_NAME_RUNES = 2;
export const MAX_DISPLAY_NAME_RUNES = 50;

/** Matches backend `MinGroupNameRunes` / `MaxGroupNameRunes`. */
export const MIN_GROUP_NAME_RUNES = 2;
export const MAX_GROUP_NAME_RUNES = 80;

/** Matches backend `username.MinLen` / `username.MaxLen`. */
export const MIN_USERNAME_LEN = 3;
export const MAX_USERNAME_LEN = 30;

export const RESERVED_USERNAMES = new Set([
	"admin",
	"administrator",
	"api",
	"auth",
	"bot",
	"channel",
	"everyone",
	"group",
	"help",
	"here",
	"invite",
	"login",
	"mod",
	"moderator",
	"null",
	"owner",
	"register",
	"root",
	"settings",
	"support",
	"system",
	"undefined",
	"wichat",
]);

export function unicodeRuneCount(value: string): number {
	let count = 0;
	for (const _ of value) {
		count++;
	}
	return count;
}

const CONTROL_OR_INVISIBLE_RE = /[\p{Cc}\p{Cf}]/u;
const CONTROL_OR_INVISIBLE_ALLOW_NEWLINE_RE = /[^\n\P{Cc}]|\p{Cf}/u;
const SPOOFING_CHARS_RE = /[@#:`]/;
const USERNAME_PATTERN =
	/^[a-z0-9](?:[a-z0-9]|(?<![._])[._](?![._]))*[a-z0-9]$/;

export function hasControlOrInvisible(
	value: string,
	allowNewline = false,
): boolean {
	if (!allowNewline) {
		return CONTROL_OR_INVISIBLE_RE.test(value);
	}
	for (const ch of value) {
		if (ch === "\n") continue;
		if (CONTROL_OR_INVISIBLE_ALLOW_NEWLINE_RE.test(ch)) {
			return true;
		}
	}
	return false;
}

export const displayNameSchema = z
	.string()
	.trim()
	.refine(
		(v) => unicodeRuneCount(v) >= MIN_DISPLAY_NAME_RUNES,
		"validation.display_name",
	)
	.refine(
		(v) => unicodeRuneCount(v) <= MAX_DISPLAY_NAME_RUNES,
		"validation.display_name_length",
	)
	.refine(
		(v) => !hasControlOrInvisible(v, false) && !SPOOFING_CHARS_RE.test(v),
		"validation.display_name_chars",
	);

export const groupNameSchema = z
	.string()
	.trim()
	.refine(
		(v) => unicodeRuneCount(v) >= MIN_GROUP_NAME_RUNES,
		"validation.group_name",
	)
	.refine(
		(v) => unicodeRuneCount(v) <= MAX_GROUP_NAME_RUNES,
		"validation.group_name_length",
	)
	.refine(
		(v) => !hasControlOrInvisible(v, false),
		"validation.group_name_chars",
	);

export const usernameSchema = z
	.string()
	.trim()
	.toLowerCase()
	.min(MIN_USERNAME_LEN, "validation.username_short")
	.max(MAX_USERNAME_LEN, "validation.username_long")
	.refine((v) => USERNAME_PATTERN.test(v), "validation.username_format")
	.refine((v) => !RESERVED_USERNAMES.has(v), "validation.username_reserved");
