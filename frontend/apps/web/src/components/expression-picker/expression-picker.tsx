"use client";

import { useEffect, useMemo, useState, useSyncExternalStore } from "react";
import { useTranslation } from "react-i18next";
import { AllTabPanel } from "@/components/expression-picker/all-tab-panel";
import { EmojiTabPanel } from "@/components/expression-picker/emoji-tab-panel";
import { ExpressionSearchBar } from "@/components/expression-picker/expression-search-bar";
import { GifsTabPanel } from "@/components/expression-picker/gifs-tab-panel";
import { StickersTabPanel } from "@/components/expression-picker/stickers-tab-panel";
import type {
	ExpressionSelection,
	ExpressionTab,
	GifEntry,
	StickerEntry,
} from "@/lib/expressions/expression-types";
import {
	getRecentEmojis,
	getRecentStickers,
	getSavedSkinTone,
	recordRecentEmoji,
	recordRecentGif,
	recordRecentSticker,
	saveSkinTone,
	subscribeRecentExpressions,
} from "@/lib/expressions/recent-expressions-storage";
import { cn } from "@/lib/utils";
import { useExpressionMediaQuery } from "@/services/expression/expression.query";

const EMPTY_STRING_LIST: string[] = [];
const EMPTY_STICKER_LIST: StickerEntry[] = [];

export type ExpressionPickerProps = {
	/** Shortcut to render a single feature mode without top tabs (e.g. "emoji"). */
	only?: "emoji" | "gifs" | "stickers";
	/** Enable/disable Emoji tab and Emoji section in All tab (default: true). */
	enableEmoji?: boolean;
	/** Enable/disable GIFs tab and GIFs section in All tab (default: true). */
	enableGifs?: boolean;
	/** Enable/disable Stickers tab and Stickers section in All tab (default: true). */
	enableStickers?: boolean;
	/** Show "All" overview tab when 2+ features are enabled (default: true). */
	enableAllTab?: boolean;
	/** Initial active tab (falls back to first available tab). */
	defaultTab?: ExpressionTab;
	/** Controlled active tab. */
	activeTab?: ExpressionTab;
	onTabChange?: (tab: ExpressionTab) => void;
	/** Called when user picks an emoji, GIF, or sticker. */
	onSelect: (selection: ExpressionSelection) => void;
	className?: string;
};

export function ExpressionPicker({
	only,
	enableEmoji = true,
	enableGifs = true,
	enableStickers = true,
	enableAllTab = true,
	defaultTab = "all",
	activeTab,
	onTabChange,
	onSelect,
	className,
}: ExpressionPickerProps) {
	const { t } = useTranslation();
	const [query, setQuery] = useState("");
	const [debouncedQuery, setDebouncedQuery] = useState("");

	const allowEmoji = only ? only === "emoji" : enableEmoji;
	const allowGifs = only ? only === "gifs" : enableGifs;
	const allowStickers = only ? only === "stickers" : enableStickers;

	useEffect(() => {
		const timer = setTimeout(() => {
			setDebouncedQuery(query);
		}, 280);
		return () => clearTimeout(timer);
	}, [query]);

	const gifsProbe = useExpressionMediaQuery({
		kind: "gifs",
		query: "",
		limit: 3,
		enabled: allowGifs,
	});
	const stickersProbe = useExpressionMediaQuery({
		kind: "stickers",
		query: "",
		limit: 4,
		enabled: allowStickers,
	});

	const gifsAvailable =
		allowGifs &&
		!gifsProbe.isError &&
		(!gifsProbe.isSuccess ||
			(gifsProbe.data.configured && gifsProbe.data.ok && gifsProbe.data.items.length > 0));

	const stickersAvailable =
		allowStickers &&
		!stickersProbe.isError &&
		(!stickersProbe.isSuccess ||
			(stickersProbe.data.configured &&
				stickersProbe.data.ok &&
				stickersProbe.data.items.length > 0));

	const enabledTabs = useMemo<ExpressionTab[]>(() => {
		const featureTabs: ExpressionTab[] = [];
		if (allowEmoji) featureTabs.push("emoji");
		if (gifsAvailable) featureTabs.push("gifs");
		if (stickersAvailable) featureTabs.push("stickers");
		if (enableAllTab && featureTabs.length >= 2) {
			return ["all", ...featureTabs];
		}
		return featureTabs;
	}, [allowEmoji, enableAllTab, gifsAvailable, stickersAvailable]);

	const fallbackTab: ExpressionTab = enabledTabs.includes(defaultTab)
		? defaultTab
		: (enabledTabs[0] ?? "emoji");

	const [internalTab, setInternalTab] = useState<ExpressionTab>(fallbackTab);
	const resolvedTab = useMemo<ExpressionTab>(() => {
		const candidate = activeTab ?? internalTab;
		return enabledTabs.includes(candidate) ? candidate : fallbackTab;
	}, [activeTab, enabledTabs, fallbackTab, internalTab]);

	const setTab = (next: ExpressionTab) => {
		if (!enabledTabs.includes(next)) {
			return;
		}
		if (activeTab === undefined) {
			setInternalTab(next);
		}
		onTabChange?.(next);
	};

	const recentEmojis = useSyncExternalStore(
		subscribeRecentExpressions,
		getRecentEmojis,
		() => EMPTY_STRING_LIST,
	);
	const recentStickers = useSyncExternalStore(
		subscribeRecentExpressions,
		getRecentStickers,
		() => EMPTY_STICKER_LIST,
	);
	const skinTone = useSyncExternalStore(
		subscribeRecentExpressions,
		getSavedSkinTone,
		() => 0,
	);

	if (enabledTabs.length === 0) {
		return null;
	}

	const handleSelectEmoji = (emoji: string, name: string) => {
		recordRecentEmoji(emoji);
		onSelect({ kind: "emoji", emoji, name });
	};

	const handleSelectGif = (gif: GifEntry) => {
		recordRecentGif(gif);
		onSelect({
			kind: "gif",
			id: gif.id,
			title: gif.title,
			url: gif.url,
			previewUrl: gif.previewUrl,
			width: gif.width,
			height: gif.height,
			provider: gif.provider,
		});
	};

	const handleSelectSticker = (sticker: StickerEntry) => {
		recordRecentSticker(sticker);
		onSelect({
			kind: "sticker",
			id: sticker.id,
			title: sticker.title,
			url: sticker.url,
			previewUrl: sticker.previewUrl,
			width: sticker.width,
			height: sticker.height,
			provider: sticker.provider,
		});
	};

	const showTabBar = enabledTabs.length > 1;

	return (
		<div
			className={cn(
				"flex w-[336px] flex-col overflow-hidden rounded-xl bg-card text-card-foreground sm:w-[360px]",
				showTabBar ? "h-[460px]" : "h-[416px]",
				className,
			)}
		>
			{showTabBar ? (
				<div
					role="tablist"
					className="flex items-center justify-around px-3 pt-2"
				>
					{enabledTabs.map((tab) => {
						const active = resolvedTab === tab;
						return (
							<button
								key={tab}
								type="button"
								role="tab"
								aria-selected={active}
								onClick={() => setTab(tab)}
								className={cn(
									"relative px-3 py-2 text-[13px] font-medium transition-colors",
									active
										? "text-foreground"
										: "text-muted-foreground hover:text-foreground",
								)}
							>
								{t(`expression.tabs.${tab}`)}
								{active ? (
									<span className="absolute inset-x-2 bottom-0 h-0.5 rounded-full bg-primary" />
								) : null}
							</button>
						);
					})}
				</div>
			) : null}

			<ExpressionSearchBar
				value={query}
				onChange={setQuery}
				showSkinTonePicker={resolvedTab === "emoji"}
				skinTone={skinTone}
				onSkinToneChange={saveSkinTone}
			/>

			<div className="flex min-h-0 flex-1 flex-col">
			{resolvedTab === "all" ? (
					<AllTabPanel
						query={query}
						debouncedQuery={debouncedQuery}
						enableEmoji={allowEmoji}
						enableGifs={gifsAvailable}
					enableStickers={stickersAvailable}
					recentEmojis={recentEmojis}
					recentStickers={recentStickers}
					onNavigateTab={setTab}
					onSelectEmoji={handleSelectEmoji}
					onSelectGif={handleSelectGif}
					onSelectSticker={handleSelectSticker}
				/>
			) : null}

			{resolvedTab === "emoji" ? (
				<EmojiTabPanel
					query={query}
					recentEmojis={recentEmojis}
					skinTone={skinTone}
					onSelectEmoji={handleSelectEmoji}
				/>
			) : null}

			{resolvedTab === "gifs" && gifsAvailable ? (
				<GifsTabPanel query={debouncedQuery} onSelectGif={handleSelectGif} />
			) : null}

			{resolvedTab === "stickers" && stickersAvailable ? (
				<StickersTabPanel
					query={debouncedQuery}
					recentStickers={recentStickers}
					onSelectSticker={handleSelectSticker}
				/>
			) : null}
			</div>
		</div>
	);
}
