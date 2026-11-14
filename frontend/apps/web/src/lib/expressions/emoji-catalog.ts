import type { EmojiCategoryId, EmojiEntry } from "@/lib/expressions/expression-types";

type RawCompactEmoji = {
	hexcode: string;
	label: string;
	unicode: string;
	group?: number;
	order?: number;
	tags?: string[];
	skins?: Array<{
		hexcode: string;
		label: string;
		unicode: string;
		tone: number | number[];
	}>;
};

const GROUP_TO_CATEGORY: Record<number, EmojiCategoryId> = {
	0: "smileys",
	1: "people",
	3: "nature",
	4: "food",
	5: "travel",
	6: "activities",
	7: "objects",
	8: "symbols",
	9: "flags",
};

export const EMOJI_CATEGORY_ORDER: EmojiCategoryId[] = [
	"smileys",
	"people",
	"nature",
	"food",
	"travel",
	"activities",
	"objects",
	"symbols",
	"flags",
];

export type EmojiCatalog = {
	all: EmojiEntry[];
	byCategory: Record<EmojiCategoryId, EmojiEntry[]>;
};

let cachedCatalog: EmojiCatalog | null = null;
let inflightLoad: Promise<EmojiCatalog> | null = null;

export async function loadEmojiCatalog(): Promise<EmojiCatalog> {
	if (cachedCatalog) {
		return cachedCatalog;
	}
	if (inflightLoad) {
		return inflightLoad;
	}

	inflightLoad = import("emojibase-data/en/compact.json").then((mod) => {
		const rawList = (mod.default ?? mod) as RawCompactEmoji[];
		const all: EmojiEntry[] = [];
		const byCategory: Record<EmojiCategoryId, EmojiEntry[]> = {
			smileys: [],
			people: [],
			nature: [],
			food: [],
			travel: [],
			activities: [],
			objects: [],
			symbols: [],
			flags: [],
		};

		for (const raw of rawList) {
			if (raw.group === undefined) {
				continue;
			}
			const category = GROUP_TO_CATEGORY[raw.group];
			if (!category) {
				continue;
			}
			const entry: EmojiEntry = {
				hexcode: raw.hexcode,
				unicode: raw.unicode,
				label: raw.label,
				category,
				tags: raw.tags ?? [],
				skins: raw.skins,
			};
			all.push(entry);
			byCategory[category].push(entry);
		}

		cachedCatalog = { all, byCategory };
		return cachedCatalog;
	});

	return inflightLoad;
}

export function getCachedEmojiCatalog(): EmojiCatalog | null {
	return cachedCatalog;
}

export function searchEmojisInCatalog(
	catalog: EmojiCatalog,
	query: string,
	limit = 98,
): EmojiEntry[] {
	const needle = query.trim().toLowerCase();
	if (!needle) {
		return catalog.all.slice(0, limit);
	}
	const results: EmojiEntry[] = [];
	for (const entry of catalog.all) {
		if (
			entry.label.toLowerCase().includes(needle) ||
			entry.tags.some((tag) => tag.toLowerCase().includes(needle))
		) {
			results.push(entry);
			if (results.length >= limit) {
				break;
			}
		}
	}
	return results;
}

const TONE_MODIFIER_HEX = ["", "1F3FB", "1F3FC", "1F3FD", "1F3FE", "1F3FF"] as const;

export function resolveEmojiWithTone(entry: EmojiEntry, skinTone: number): string {
	const modifier = TONE_MODIFIER_HEX[skinTone];
	if (!modifier || !entry.skins || entry.skins.length === 0) {
		return entry.unicode;
	}
	const otherModifiers = TONE_MODIFIER_HEX.filter((hex) => hex && hex !== modifier);
	const match = entry.skins.find(
		(skin) =>
			skin.hexcode.includes(modifier) &&
			!otherModifiers.some((other) => skin.hexcode.includes(other)),
	);
	return match?.unicode ?? entry.skins[skinTone - 1]?.unicode ?? entry.unicode;
}

export const DEFAULT_QUICK_EMOJIS: string[] = [
	"😀",
	"😂",
	"🥹",
	"🙏",
	"🤣",
	"😅",
	"🔥",
	"❤️",
	"👍",
	"🎉",
	"💯",
	"👀",
];
