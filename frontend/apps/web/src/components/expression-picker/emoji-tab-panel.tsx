"use client";

import {
	Building2,
	Cat,
	Clock,
	Flag,
	Hand,
	Hash,
	Lightbulb,
	Loader2,
	Pizza,
	Smile,
	Trophy,
} from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
	EMOJI_CATEGORY_ORDER,
	getCachedEmojiCatalog,
	loadEmojiCatalog,
	resolveEmojiWithTone,
	searchEmojisInCatalog,
	type EmojiCatalog,
} from "@/lib/expressions/emoji-catalog";
import type { EmojiCategoryId, EmojiEntry } from "@/lib/expressions/expression-types";
import { cn } from "@/lib/utils";

const CATEGORY_ICONS: Record<EmojiCategoryId, typeof Smile> = {
	smileys: Smile,
	people: Hand,
	nature: Cat,
	food: Pizza,
	travel: Building2,
	activities: Trophy,
	objects: Lightbulb,
	symbols: Hash,
	flags: Flag,
};

const COLUMNS = 7;
const HEADER_HEIGHT = 30;
const ROW_HEIGHT = 42;
const VIEWPORT_HEIGHT = 296;
const OVERSCAN_PX = 160;

type VirtualItem =
	| {
			key: string;
			kind: "header";
			label: string;
			top: number;
			height: number;
	  }
	| {
			key: string;
			kind: "recent-row";
			emojis: string[];
			top: number;
			height: number;
	  }
	| {
			key: string;
			kind: "emoji-row";
			entries: EmojiEntry[];
			top: number;
			height: number;
	  };

type EmojiTabPanelProps = {
	query: string;
	recentEmojis: string[];
	skinTone: number;
	onSelectEmoji: (emoji: string, name: string) => void;
};

export function EmojiTabPanel({
	query,
	recentEmojis,
	skinTone,
	onSelectEmoji,
}: EmojiTabPanelProps) {
	const { t } = useTranslation();
	const scrollRef = useRef<HTMLDivElement>(null);
	const [scrollTop, setScrollTop] = useState(0);
	const [catalog, setCatalog] = useState<EmojiCatalog | null>(() => getCachedEmojiCatalog());

	useEffect(() => {
		if (catalog) {
			return;
		}
		let cancelled = false;
		void loadEmojiCatalog().then((loaded) => {
			if (!cancelled) {
				setCatalog(loaded);
			}
		});
		return () => {
			cancelled = true;
		};
	}, [catalog]);

	const trimmedQuery = query.trim();

	const searchResults = useMemo(() => {
		if (!trimmedQuery || !catalog) {
			return [];
		}
		return searchEmojisInCatalog(catalog, trimmedQuery, 98);
	}, [catalog, trimmedQuery]);

	const { items, totalHeight, sectionOffsets } = useMemo(() => {
		if (!catalog) {
			return {
				items: [] as VirtualItem[],
				totalHeight: 0,
				sectionOffsets: {} as Record<string, number>,
			};
		}

		const list: VirtualItem[] = [];
		const offsets: Record<string, number> = {};
		let y = 0;

		if (recentEmojis.length > 0) {
			offsets.recent = y;
			list.push({
				key: "header-recent",
				kind: "header",
				label: t("expression.recent"),
				top: y,
				height: HEADER_HEIGHT,
			});
			y += HEADER_HEIGHT;

			for (let i = 0; i < recentEmojis.length; i += COLUMNS) {
				list.push({
					key: `recent-row-${i}`,
					kind: "recent-row",
					emojis: recentEmojis.slice(i, i + COLUMNS),
					top: y,
					height: ROW_HEIGHT,
				});
				y += ROW_HEIGHT;
			}
		}

		for (const categoryId of EMOJI_CATEGORY_ORDER) {
			const categoryEmojis = catalog.byCategory[categoryId];
			if (categoryEmojis.length === 0) {
				continue;
			}
			offsets[categoryId] = y;
			list.push({
				key: `header-${categoryId}`,
				kind: "header",
				label: t(`expression.emoji_categories.${categoryId}`),
				top: y,
				height: HEADER_HEIGHT,
			});
			y += HEADER_HEIGHT;

			for (let i = 0; i < categoryEmojis.length; i += COLUMNS) {
				list.push({
					key: `${categoryId}-row-${i}`,
					kind: "emoji-row",
					entries: categoryEmojis.slice(i, i + COLUMNS),
					top: y,
					height: ROW_HEIGHT,
				});
				y += ROW_HEIGHT;
			}
		}

		return { items: list, totalHeight: y, sectionOffsets: offsets };
	}, [catalog, recentEmojis, t]);

	const visibleItems = useMemo(() => {
		const minTop = scrollTop - OVERSCAN_PX;
		const maxBottom = scrollTop + VIEWPORT_HEIGHT + OVERSCAN_PX;
		return items.filter(
			(item) => item.top + item.height >= minTop && item.top <= maxBottom,
		);
	}, [items, scrollTop]);

	const scrollToSection = (key: string) => {
		const container = scrollRef.current;
		const offset = sectionOffsets[key];
		if (!container || offset === undefined) {
			return;
		}
		container.scrollTo({ top: offset, behavior: "smooth" });
	};

	if (!catalog) {
		return (
			<div className="flex h-full items-center justify-center">
				<Loader2 className="size-5 animate-spin text-muted-foreground" />
			</div>
		);
	}

	if (trimmedQuery) {
		return (
			<div className="thin-scroll h-full min-h-0 overflow-y-auto px-5 pb-3.5">
				{searchResults.length === 0 ? (
					<div className="flex h-full items-center justify-center text-xs text-muted-foreground">
						{t("expression.no_results")}
					</div>
				) : (
					<div className="grid grid-cols-7 gap-1">
						{searchResults.map((entry) => {
							const glyph = resolveEmojiWithTone(entry, skinTone);
							return (
								<button
									key={entry.hexcode}
									type="button"
									title={entry.label}
									aria-label={entry.label}
									onClick={() => onSelectEmoji(glyph, entry.label)}
									className="flex aspect-square items-center justify-center rounded-md text-[22px] leading-none transition-colors hover:bg-accent"
								>
									{glyph}
								</button>
							);
						})}
					</div>
				)}
			</div>
		);
	}

	return (
		<div className="flex h-full min-h-0 flex-col">
			<div
				ref={scrollRef}
				onScroll={(event) => setScrollTop(event.currentTarget.scrollTop)}
				className="thin-scroll min-h-0 flex-1 overflow-y-auto px-5 pb-2.5"
			>
				<div className="relative w-full" style={{ height: totalHeight }}>
					{visibleItems.map((item) => {
						if (item.kind === "header") {
							return (
								<div
									key={item.key}
									style={{ top: item.top, height: item.height }}
									className="absolute inset-x-0 flex items-center text-xs font-semibold text-foreground"
								>
									{item.label}
								</div>
							);
						}

						if (item.kind === "recent-row") {
							return (
								<div
									key={item.key}
									style={{ top: item.top, height: item.height }}
									className="absolute inset-x-0 grid grid-cols-7 gap-1"
								>
									{item.emojis.map((emoji) => (
										<button
											key={`recent-${emoji}`}
											type="button"
											aria-label={emoji}
											onClick={() => onSelectEmoji(emoji, emoji)}
											className="flex h-9 items-center justify-center rounded-md text-[22px] leading-none transition-colors hover:bg-accent"
										>
											{emoji}
										</button>
									))}
								</div>
							);
						}

						return (
							<div
								key={item.key}
								style={{ top: item.top, height: item.height }}
								className="absolute inset-x-0 grid grid-cols-7 gap-1"
							>
								{item.entries.map((entry) => {
									const glyph = resolveEmojiWithTone(entry, skinTone);
									return (
										<button
											key={entry.hexcode}
											type="button"
											title={entry.label}
											aria-label={entry.label}
											onClick={() => onSelectEmoji(glyph, entry.label)}
											className="flex h-9 items-center justify-center rounded-md text-[22px] leading-none transition-colors hover:bg-accent"
										>
											{glyph}
										</button>
									);
								})}
							</div>
						);
					})}
				</div>
			</div>

			<div className="flex items-center justify-between border-t border-border/70 px-3 pt-1.5 pb-4">
				{recentEmojis.length > 0 ? (
					<button
						type="button"
						onClick={() => scrollToSection("recent")}
						title={t("expression.emoji_categories.recent")}
						aria-label={t("expression.emoji_categories.recent")}
						className="flex size-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
					>
						<Clock className="size-4" />
					</button>
				) : null}
				{EMOJI_CATEGORY_ORDER.map((categoryId) => {
					const Icon = CATEGORY_ICONS[categoryId];
					return (
						<button
							key={`nav-${categoryId}`}
							type="button"
							onClick={() => scrollToSection(categoryId)}
							title={t(`expression.emoji_categories.${categoryId}`)}
							aria-label={t(`expression.emoji_categories.${categoryId}`)}
							className={cn(
								"flex size-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground",
							)}
						>
							<Icon className="size-4" />
						</button>
					);
				})}
			</div>
		</div>
	);
}
