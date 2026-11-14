"use client";

import { ChevronRight } from "lucide-react";
import Image from "next/image";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
	DEFAULT_QUICK_EMOJIS,
	getCachedEmojiCatalog,
	loadEmojiCatalog,
	searchEmojisInCatalog,
	type EmojiCatalog,
} from "@/lib/expressions/emoji-catalog";
import type {
	ExpressionTab,
	GifEntry,
	StickerEntry,
} from "@/lib/expressions/expression-types";
import { useExpressionMediaQuery } from "@/services/expression/expression.query";

type AllTabPanelProps = {
	query: string;
	debouncedQuery: string;
	enableEmoji: boolean;
	enableGifs: boolean;
	enableStickers: boolean;
	recentEmojis: string[];
	recentStickers: StickerEntry[];
	onNavigateTab: (tab: ExpressionTab) => void;
	onSelectEmoji: (emoji: string, name: string) => void;
	onSelectGif: (gif: GifEntry) => void;
	onSelectSticker: (sticker: StickerEntry) => void;
};

export function AllTabPanel({
	query,
	debouncedQuery,
	enableEmoji,
	enableGifs,
	enableStickers,
	recentEmojis,
	recentStickers,
	onNavigateTab,
	onSelectEmoji,
	onSelectGif,
	onSelectSticker,
}: AllTabPanelProps) {
	const { t } = useTranslation();
	const trimmedQuery = query.trim();
	const [catalog, setCatalog] = useState<EmojiCatalog | null>(() => getCachedEmojiCatalog());

	useEffect(() => {
		if (!enableEmoji || !trimmedQuery || catalog) {
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
	}, [catalog, enableEmoji, trimmedQuery]);

	const emojiPreview = useMemo(() => {
		if (!enableEmoji) {
			return [];
		}
		if (trimmedQuery) {
			if (!catalog) {
				return [];
			}
			return searchEmojisInCatalog(catalog, trimmedQuery, 12).map((entry) => ({
				emoji: entry.unicode,
				label: entry.label,
			}));
		}
		const source = recentEmojis.length > 0 ? recentEmojis : DEFAULT_QUICK_EMOJIS;
		return source.slice(0, 6).map((emoji) => ({ emoji, label: emoji }));
	}, [catalog, enableEmoji, recentEmojis, trimmedQuery]);

	const gifsQuery = useExpressionMediaQuery({
		kind: "gifs",
		query: debouncedQuery,
		limit: 3,
		enabled: enableGifs,
	});

	const stickersQuery = useExpressionMediaQuery({
		kind: "stickers",
		query: debouncedQuery,
		limit: 4,
		enabled: enableStickers,
	});

	const gifPreview = gifsQuery.data?.items ?? [];
	const gifAvailable =
		enableGifs &&
		!gifsQuery.isError &&
		Boolean(gifsQuery.data?.configured && gifsQuery.data?.ok && gifPreview.length > 0);

	const apiStickers = stickersQuery.data?.items ?? [];
	const stickerPreview =
		!debouncedQuery.trim() && recentStickers.length > 0 && apiStickers.length > 0
			? recentStickers.slice(0, 4)
			: apiStickers.slice(0, 4);
	const stickerAvailable =
		enableStickers &&
		!stickersQuery.isError &&
		Boolean(
			stickersQuery.data?.configured &&
				stickersQuery.data?.ok &&
				stickerPreview.length > 0,
		);

	const activeProvider =
		(gifAvailable ? gifsQuery.data?.provider : undefined) ??
		(stickerAvailable ? stickersQuery.data?.provider : undefined) ??
		"none";

	return (
		<div className="flex h-full min-h-0 flex-col">
			<div className="thin-scroll min-h-0 flex-1 space-y-5 overflow-y-auto px-5 pb-3.5">
				{enableEmoji && emojiPreview.length > 0 ? (
					<section>
						<div className="mb-2 flex items-center justify-between">
							<h3 className="text-xs font-semibold text-foreground">
								{t("expression.tabs.emoji")}
							</h3>
							<button
								type="button"
								onClick={() => onNavigateTab("emoji")}
								className="inline-flex items-center gap-0.5 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
							>
								<span>{t("expression.see_all")}</span>
								<ChevronRight className="size-3.5" />
							</button>
						</div>
						<div className="grid grid-cols-6 gap-1">
							{emojiPreview.map((item) => (
								<button
									key={`all-emoji-${item.emoji}`}
									type="button"
									title={item.label}
									aria-label={item.label}
									onClick={() => onSelectEmoji(item.emoji, item.label)}
									className="flex aspect-square items-center justify-center rounded-md text-[24px] leading-none transition-colors hover:bg-accent"
								>
									{item.emoji}
								</button>
							))}
						</div>
					</section>
				) : null}

				{gifAvailable ? (
					<section>
						<div className="mb-1.5 flex items-center justify-between">
							<h3 className="text-xs font-semibold text-foreground">
								{t("expression.tabs.gifs")}
							</h3>
							<button
								type="button"
								onClick={() => onNavigateTab("gifs")}
								className="inline-flex items-center gap-0.5 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
							>
								<span>{t("expression.see_all")}</span>
								<ChevronRight className="size-3.5" />
							</button>
						</div>
						<div className="grid grid-cols-3 gap-1.5">
							{gifPreview.slice(0, 3).map((gif) => (
								<button
									key={`all-gif-${gif.id}`}
									type="button"
									title={gif.title}
									onClick={() => onSelectGif(gif)}
									className="group relative aspect-[4/3] overflow-hidden rounded-lg border border-border/50 bg-muted"
								>
									<Image
										src={gif.previewUrl}
										alt={gif.title}
										fill
										unoptimized
										loading="lazy"
										decoding="async"
										className="object-cover transition-transform duration-150 group-hover:scale-105"
									/>
								</button>
							))}
						</div>
					</section>
				) : null}

				{stickerAvailable ? (
					<section>
						<div className="mb-1.5 flex items-center justify-between">
							<h3 className="text-xs font-semibold text-foreground">
								{t("expression.tabs.stickers")}
							</h3>
							<button
								type="button"
								onClick={() => onNavigateTab("stickers")}
								className="inline-flex items-center gap-0.5 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
							>
								<span>{t("expression.see_all")}</span>
								<ChevronRight className="size-3.5" />
							</button>
						</div>
						<div className="grid grid-cols-4 gap-1.5">
							{stickerPreview.map((sticker) => (
								<button
									key={`all-sticker-${sticker.id}`}
									type="button"
									title={sticker.title}
									aria-label={sticker.title}
									onClick={() => onSelectSticker(sticker)}
									className="flex aspect-square items-center justify-center overflow-hidden rounded-lg bg-[#f5f5f6]/70 p-1.5 transition-colors hover:bg-accent dark:bg-[#242424]/70 dark:hover:bg-accent"
								>
									<div className="relative size-full">
										<Image
											src={sticker.previewUrl}
											alt={sticker.title}
											fill
											unoptimized
											loading="lazy"
											decoding="async"
											className="object-contain"
										/>
									</div>
								</button>
							))}
						</div>
					</section>
				) : null}

				{emojiPreview.length === 0 && !gifAvailable && !stickerAvailable ? (
					<div className="flex h-40 items-center justify-center text-xs text-muted-foreground">
						{t("expression.no_results")}
					</div>
				) : null}
			</div>

			{activeProvider !== "none" ? (
				<div className="flex items-center justify-end px-5 pt-1.5 pb-4">
					<span className="text-[11px] text-muted-foreground">
						{t("expression.gifs_powered_by", {
							provider: activeProvider === "klipy" ? "KLIPY" : "GIPHY",
						})}
					</span>
				</div>
			) : null}
		</div>
	);
}
