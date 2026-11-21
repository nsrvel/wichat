"use client";

import type { InvitePreview } from "@wichat/contracts/group/invite";
import Image from "next/image";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { GroupMemberStatsPill } from "@/components/group/group-member-stats-pill";
import { Button } from "@/components/ui/button";
import { groupBannerSwatchSurfaceClass } from "@/lib/group-banner-presets";
import { authSubmitClass } from "@/lib/form-field-styles";
import { groupInitials } from "@/lib/group-initials";
import { resolveGroupLogoUrl } from "@/lib/group-logo-url";
import { cn } from "@/lib/utils";

type InviteCardBannerProps = {
	preview: InvitePreview;
};

/**
 * Full-bleed banner header for the invite acceptance card, rendering the group's
 * custom banner image or banner color with a clean fallback.
 */
export function InviteCardBanner({ preview }: InviteCardBannerProps) {
	const [bannerBroken, setBannerBroken] = useState(false);
	if (preview.is_private) {
		return null;
	}

	const bannerSrc = resolveGroupLogoUrl(preview.group_banner_url ?? "");
	const showBannerImage = Boolean(bannerSrc) && !bannerBroken;
	const rawBannerColor = preview.group_banner_color?.trim() ?? "";
	const isCustomHexBanner = /^#[0-9a-f]{3,8}$/i.test(rawBannerColor);
	const presetBannerClass =
		!showBannerImage && rawBannerColor && !isCustomHexBanner
			? groupBannerSwatchSurfaceClass(rawBannerColor)
			: "";

	return (
		<div
			className={cn(
				"relative h-28 w-full overflow-hidden border-b border-border/60 bg-muted/60 sm:h-32",
				presetBannerClass,
			)}
			style={
				!showBannerImage && !presetBannerClass
					? {
							background: isCustomHexBanner
								? `linear-gradient(135deg, ${rawBannerColor} 0%, color-mix(in oklab, ${rawBannerColor} 68%, var(--card)) 100%)`
								: "linear-gradient(135deg, color-mix(in oklab, var(--primary) 16%, var(--card)) 0%, var(--muted) 100%)",
						}
					: undefined
			}
		>
			{showBannerImage ? (
				<Image
					src={bannerSrc}
					alt=""
					fill
					unoptimized
					loading="eager"
					fetchPriority="high"
					sizes="(max-width: 640px) 100vw, 448px"
					className="object-cover"
					onError={() => setBannerBroken(true)}
				/>
			) : null}
		</div>
	);
}

type InviteAcceptCardProps = {
	preview: InvitePreview;
	acceptPending?: boolean;
	acceptDisabled?: boolean;
	hasBanner?: boolean;
	onAccept: () => void;
	className?: string;
};

export function InviteAcceptCard({
	preview,
	acceptPending = false,
	acceptDisabled = false,
	hasBanner = false,
	onAccept,
	className,
}: InviteAcceptCardProps) {
	const { t } = useTranslation();
	const [logoBroken, setLogoBroken] = useState(false);
	const isPrivate = Boolean(preview.is_private);
	const logoSrc = isPrivate ? "" : resolveGroupLogoUrl(preview.group_logo_url ?? "");
	const showLogo = Boolean(logoSrc) && !logoBroken;
	const memberCount = preview.member_count ?? 0;
	const onlineCount = preview.online_count ?? 0;
	const groupDescription = isPrivate ? "" : (preview.group_description?.trim() ?? "");

	return (
		<div className={cn("flex w-full flex-col items-center text-center", className)}>
			<div
				className={cn(
					"relative z-10 flex size-[4.5rem] shrink-0 items-center justify-center overflow-hidden rounded-2xl border-4 border-card bg-card shadow-md sm:size-20",
					hasBanner && !isPrivate ? "-mt-9 mb-3 sm:-mt-10 sm:mb-3.5" : "mb-4 sm:mb-5",
				)}
			>
				{showLogo ? (
					<Image
						src={logoSrc}
						alt=""
						width={80}
						height={80}
						unoptimized
						loading="eager"
						fetchPriority="high"
						decoding="async"
						className="size-full object-cover"
						onError={() => setLogoBroken(true)}
					/>
				) : (
					<span className="flex size-full items-center justify-center bg-primary/15 text-lg font-semibold text-primary sm:text-xl">
						{groupInitials(preview.group_name ?? "")}
					</span>
				)}
			</div>

			<p className="mb-1 text-xs font-medium text-muted-foreground sm:text-[13px]">
				{t("invite.landing_eyebrow")}
			</p>
			<h2 className="mb-2.5 w-full min-w-0 truncate text-xl font-bold tracking-tight text-foreground sm:text-2xl">
				{preview.group_name}
			</h2>

			<GroupMemberStatsPill onlineCount={onlineCount} memberCount={memberCount} className="mb-4" />

			{groupDescription ? (
				<p className="mb-4 line-clamp-3 max-w-xs whitespace-pre-wrap break-words text-xs leading-relaxed text-muted-foreground sm:text-[13px]">
					{groupDescription}
				</p>
			) : null}

			{preview.channel_name ? (
				<p className="mb-4 text-xs text-muted-foreground sm:mb-5">
					{t("invite.landing_lead_channel", { channel: preview.channel_name })}
				</p>
			) : !groupDescription ? (
				<div className="mb-1" />
			) : null}

			<Button
				type="button"
				disabled={acceptPending || acceptDisabled}
				className={cn("w-full", authSubmitClass)}
				onClick={onAccept}
			>
				{acceptPending ? t("invite.landing_joining") : t("invite.landing_accept_invite")}
			</Button>
		</div>
	);
}
