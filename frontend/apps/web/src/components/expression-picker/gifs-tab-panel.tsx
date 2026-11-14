"use client";

import { ArrowLeft, Film, Loader2 } from "lucide-react";
import Image from "next/image";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import type { GifEntry } from "@/lib/expressions/expression-types";
import {
	useExpressionCategoriesQuery,
	useExpressionMediaQuery,
} from "@/services/expression/expression.query";

type GifsTabPanelProps = {
	query: string;
	onSelectGif: (gif: GifEntry) => void;
};

export function GifsTabPanel({ query, onSelectGif }: GifsTabPanelProps) {
	const { t } = useTranslation();
	const [selectedCategory, setSelectedCategory] = useState<{
		label: string;
		query: string;
	} | null>(null);

	const activeQuery = query.trim() || selectedCategory?.query || "";
	const showCategoryCards = !query.trim() && !selectedCategory;

	const categoriesQuery = useExpressionCategoriesQuery("gifs", showCategoryCards);
	const gifsQuery = useExpressionMediaQuery({
		kind: "gifs",
		query: activeQuery,
		limit: 24,
		enabled: !showCategoryCards,
	});

	const provider =
		(showCategoryCards ? categoriesQuery.data?.provider : gifsQuery.data?.provider) ?? "none";
	const configured =
		(showCategoryCards ? categoriesQuery.data?.configured : gifsQuery.data?.configured) ?? true;

	return (
		<div className="flex h-full min-h-0 flex-col">
			{selectedCategory && !query.trim() ? (
				<div className="flex items-center gap-2 px-5 pb-1.5">
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

			<div className="thin-scroll min-h-0 flex-1 overflow-y-auto px-5 pb-2.5">
				{!configured ? (
					<div className="flex h-full flex-col items-center justify-center gap-2 px-4 text-center">
						<Film className="size-7 text-muted-foreground/70" />
						<p className="text-xs leading-relaxed text-muted-foreground">
							{t("expression.gifs_unconfigured")}
						</p>
					</div>
				) : showCategoryCards ? (
					categoriesQuery.isLoading ? (
						<div className="flex h-full items-center justify-center">
							<Loader2 className="size-5 animate-spin text-muted-foreground" />
						</div>
					) : (
						<div className="grid grid-cols-2 gap-2.5">
							{(categoriesQuery.data?.categories ?? []).map((category) => (
								<button
									key={category.id}
									type="button"
									onClick={() =>
										setSelectedCategory({
											label: category.label,
											query: category.query,
										})
									}
									className="group relative h-[68px] overflow-hidden rounded-lg border border-border/60 bg-muted text-left"
								>
									{category.previewUrl ? (
										<Image
											src={category.previewUrl}
											alt={category.label}
											fill
											unoptimized
											className="object-cover transition-transform duration-200 group-hover:scale-105"
										/>
									) : null}
									<div className="absolute inset-0 bg-gradient-to-t from-black/75 via-black/35 to-black/15" />
									<span className="relative z-10 flex h-full items-end px-2.5 pb-2 text-xs font-semibold text-white">
										{category.label}
									</span>
								</button>
							))}
						</div>
					)
				) : gifsQuery.isLoading ? (
					<div className="flex h-full items-center justify-center">
						<Loader2 className="size-5 animate-spin text-muted-foreground" />
					</div>
				) : (gifsQuery.data?.items.length ?? 0) === 0 ? (
					<div className="flex h-full items-center justify-center text-xs text-muted-foreground">
						{t("expression.no_results")}
					</div>
				) : (
					<div className="grid grid-cols-2 gap-2">
						{(gifsQuery.data?.items ?? []).map((gif) => (
							<button
								key={gif.id}
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
									className="object-cover transition-transform duration-150 group-hover:scale-105"
								/>
							</button>
						))}
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

