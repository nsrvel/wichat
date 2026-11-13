"use client";

import type { ReactNode } from "react";
import Image from "next/image";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { GroupBannerPreviewSurface } from "@/features/group/components/group-banner-preview-surface";
import {
	groupProfileCardBodyClass,
	groupProfileCardIconClass,
	groupProfileCardShellClass,
	type GroupBannerMode,
} from "@/features/group/lib/group-banner-presets";
import { groupInitials } from "@/features/shell/lib/group-initials";
import { resolveGroupLogoUrl } from "@/features/shell/lib/group-logo-url";
import { cn } from "@/lib/utils";

type GroupProfileCardPreviewProps = {
	groupName: string;
	bannerMode: GroupBannerMode;
	bannerColor: string;
	bannerImageSrc?: string;
	showBannerImage: boolean;
	logoUrl?: string;
	iconFile?: File | null;
	onBannerClick?: () => void;
	bannerOverlay?: ReactNode;
	className?: string;
};

export function GroupProfileCardPreview({
	groupName,
	bannerMode,
	bannerColor,
	bannerImageSrc,
	showBannerImage,
	logoUrl = "",
	iconFile = null,
	onBannerClick,
	bannerOverlay,
	className,
}: GroupProfileCardPreviewProps) {
	const { t } = useTranslation();
	const [brokenLogoKey, setBrokenLogoKey] = useState<string | null>(null);

	const iconBlobUrl = useMemo(() => (iconFile ? URL.createObjectURL(iconFile) : null), [iconFile]);

	useEffect(() => {
		if (!iconBlobUrl) {
			return;
		}
		return () => URL.revokeObjectURL(iconBlobUrl);
	}, [iconBlobUrl]);

	const currentLogoKey = `${logoUrl}::${iconBlobUrl ?? ""}`;
	const logoBroken = brokenLogoKey === currentLogoKey;
	const resolvedLogo = resolveGroupLogoUrl(logoUrl);
	const iconSrc = iconBlobUrl ?? (resolvedLogo && !logoBroken ? resolvedLogo : "");
	const showIconImage = Boolean(iconSrc);
	const displayName = groupName.trim() || t("shell.unknown_group");
	const initials = groupInitials(displayName);

	return (
		<div className={cn(groupProfileCardShellClass, className)}>
			<GroupBannerPreviewSurface
				bannerMode={bannerMode}
				bannerColor={bannerColor}
				bannerImageSrc={bannerImageSrc}
				showBannerImage={showBannerImage}
				onBannerClick={onBannerClick}
				bannerOverlay={bannerOverlay}
				className="max-w-none border-0"
			/>

			<div className={groupProfileCardBodyClass}>
				<div className={groupProfileCardIconClass}>
					{showIconImage ? (
						<Image
							src={iconSrc}
							alt=""
							width={56}
							height={56}
							unoptimized
							className="size-full object-cover"
							onError={() => setBrokenLogoKey(currentLogoKey)}
						/>
					) : (
						<span className="flex size-full items-center justify-center bg-[color-mix(in_oklch,var(--primary)_12%,var(--field))] text-xs font-semibold text-foreground/80">
							{initials || "?"}
						</span>
					)}
				</div>

				<p className="mt-2 w-full truncate text-sm font-semibold leading-snug text-foreground">
					{displayName}
				</p>
				<p className="mt-1 text-[0.6875rem] leading-snug text-muted-foreground">
					{t("group_settings.profile.card_preview_subtitle")}
				</p>
			</div>
		</div>
	);
}
