"use client";

import type { ReactNode } from "react";
import { Upload } from "lucide-react";
import Image from "next/image";
import { useTranslation } from "react-i18next";
import {
	groupBannerPreviewStripClass,
	groupBannerSwatchSurfaceClass,
	isGroupBannerColorDefault,
	type GroupBannerMode,
} from "@/features/group/lib/group-banner-presets";
import { cn } from "@/lib/utils";

type GroupBannerPreviewSurfaceProps = {
	bannerMode: GroupBannerMode;
	bannerColor: string;
	bannerImageSrc?: string;
	showBannerImage: boolean;
	onBannerClick?: () => void;
	bannerOverlay?: ReactNode;
	className?: string;
};

export function GroupBannerPreviewSurface({
	bannerMode,
	bannerColor,
	bannerImageSrc,
	showBannerImage,
	onBannerClick,
	bannerOverlay,
	className,
}: GroupBannerPreviewSurfaceProps) {
	const { t } = useTranslation();

	return (
		<div className={cn(groupBannerPreviewStripClass, className)}>
			{bannerMode === "color" ? (
				isGroupBannerColorDefault(bannerColor) ? (
					<div
						className={cn(
							"absolute inset-0 bg-gradient-to-br from-muted via-muted/90 to-muted-foreground/25",
						)}
						aria-hidden
					/>
				) : (
					<div
						className={cn("absolute inset-0", groupBannerSwatchSurfaceClass(bannerColor))}
						aria-hidden
					/>
				)
			) : showBannerImage && bannerImageSrc ? (
				<Image
					src={bannerImageSrc}
					alt=""
					fill
					unoptimized
					sizes="100vw"
					className="object-cover"
				/>
			) : (
				<button
					type="button"
					className={cn(
						"group/banner absolute inset-0 flex flex-col items-center justify-center gap-1.5",
						"bg-muted/40 text-muted-foreground transition-colors hover:bg-muted/55",
					)}
					onClick={onBannerClick}
				>
					<Upload className="size-4 opacity-70" strokeWidth={1.75} aria-hidden />
					<span className="px-2 text-[0.625rem] font-medium leading-tight">
						{t("group_settings.profile.banner_upload")}
					</span>
				</button>
			)}
			{bannerOverlay}
		</div>
	);
}
