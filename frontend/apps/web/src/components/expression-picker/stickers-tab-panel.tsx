"use client";

import { ArrowLeft, Loader2 } from "lucide-react";
import Image from "next/image";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import type { StickerEntry } from "@/lib/expressions/expression-types";
import {
	useExpressionCategoriesQuery,
	useExpressionMediaQuery,
} from "@/services/expression/expression.query";

type StickersTabPanelProps = {
	query: string;
	recentStickers: StickerEntry[];
	onSelectSticker: (sticker: StickerEntry) => void;
};

export function StickersTabPanel({
	query,
	recentStickers,
	onSelectSticker,
}: StickersTabPanelProps) {
	const { t } = useTranslation();
	const [selectedCategory, setSelectedCategory] = useState<{
		label: string;
		query: string;
	} | null>(null);

	const trimmedQuery = query.trim();
	const activeQuery = trimmedQuery || selectedCategory?.query || "";
	const showOverview = !trimmedQuery && !selectedCategory;

	const categoriesQuery = useExpressionCategoriesQuery("stickers", showOverview);
	const trendingPreviewQuery = useExpressionMediaQuery({
		kind: "stickers",
		query: "",
		limit: 4,
		enabled: showOverview && recentStickers.length === 0,
	});
	const stickersQuery = useExpressionMediaQuery({
		kind: "stickers",
		query: activeQuery,
		limit: 24,
		enabled: !showOverview,
	});

	const provider =
		(showOverview ? categoriesQuery.data?.provider : stickersQuery.data?.provider) ?? "none";

	const topStickers =
		recentStickers.length > 0
			? recentStickers.slice(0, 4)
			: (trendingPreviewQuery.data?.items ?? []).slice(0, 4);

	const renderStickerTile = (sticker: StickerEntry) => (
		<button
			key={sticker.id}
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
	);

	return (
		<div className="flex h-full min-h-0 flex-col">
			{selectedCategory && !trimmedQuery ? (
				<div className="flex items-center gap-2 px-5 pb-2">
					<button
						type="button"
						onClick={() => setSelectedCategory(null)}
						className="inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
					>
						<ArrowLeft className="size-3.5" />
						<span>{selectedCategory.label}</span>
					</button>
				</div>
			) : null}

			<div className="thin-scroll min-h-0 flex-1 space-y-5 overflow-y-auto px-5 pb-2.5">
				{showOverview ? (
					<>
						{topStickers.length > 0 ? (
							<div>
								<p className="mb-2 text-xs font-semibold text-foreground">
									{recentStickers.length > 0
										? t("expression.recent")
										: "Popular"}
								</p>
								<div className="grid grid-cols-4 gap-2">
									{topStickers.map(renderStickerTile)}
								</div>
							</div>
						) : null}

						<div>
							<p className="mb-2 text-xs font-semibold text-foreground">
								{t("expression.sticker_packs")}
							</p>
							{categoriesQuery.isLoading ? (
								<div className="flex h-32 items-center justify-center">
									<Loader2 className="size-5 animate-spin text-muted-foreground" />
								</div>
							) : (
								<div className="grid grid-cols-2 gap-2">
									{(categoriesQuery.data?.categories ?? []).map((pack) => (
										<button
											key={pack.id}
											type="button"
											onClick={() =>
												setSelectedCategory({
													label: pack.label,
													query: pack.query,
												})
											}
											className="flex items-center justify-between rounded-lg bg-[#f5f5f6]/70 px-3 py-2 text-left transition-colors hover:bg-accent dark:bg-[#242424]/70 dark:hover:bg-accent"
										>
											<span className="truncate text-xs font-semibold text-foreground">
												{pack.label}
											</span>
											{pack.previewUrl ? (
												<div className="relative size-7 shrink-0 overflow-hidden rounded-md">
													<Image
														src={pack.previewUrl}
														alt=""
														fill
														unoptimized
														loading="lazy"
														decoding="async"
														className="object-contain"
													/>
												</div>
											) : null}
										</button>
									))}
								</div>
							)}
						</div>
					</>
				) : stickersQuery.isLoading ? (
					<div className="flex h-full items-center justify-center">
						<Loader2 className="size-5 animate-spin text-muted-foreground" />
					</div>
				) : (stickersQuery.data?.items.length ?? 0) === 0 ? (
					<div className="flex h-full items-center justify-center text-xs text-muted-foreground">
						{t("expression.no_results")}
					</div>
				) : (
					<div className="grid grid-cols-4 gap-2">
						{(stickersQuery.data?.items ?? []).map(renderStickerTile)}
					</div>
				)}
			</div>

			{provider !== "none" ? (
				<div className="flex items-center justify-end px-5 pt-1.5 pb-4">
					<span className="text-[11px] text-muted-foreground">
						{t("expression.gifs_powered_by", {
							provider: provider === "klipy" ? "KLIPY" : "GIPHY",
						})}
					</span>
				</div>
			) : null}
		</div>
	);
}
