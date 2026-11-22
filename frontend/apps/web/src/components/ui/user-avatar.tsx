"use client";

import { memo, useState } from "react";
import { resolvePublicObjectUrl, toSameOriginObjectPath } from "@/lib/public-object-url";
import { cn } from "@/lib/utils";
import { USER_STATUS_BY_ID, type UserPresenceStatus } from "@/lib/user-status";

export type UserAvatarSize = "xs" | "sm" | "md" | "lg" | "xl" | "2xl";

const SIZE_STYLES: Record<
	UserAvatarSize,
	{
		px: number;
		avatarClass: string;
		textClass: string;
		indicatorClass: string;
	}
> = {
	xs: {
		px: 24,
		avatarClass: "size-6",
		textClass: "text-[10px]",
		indicatorClass: "bottom-0 right-0 size-2 ring-[1.5px]",
	},
	sm: {
		px: 32,
		avatarClass: "size-8",
		textClass: "text-xs",
		indicatorClass: "bottom-0 right-0 size-2.5 ring-2",
	},
	md: {
		px: 36,
		avatarClass: "size-9",
		textClass: "text-xs",
		indicatorClass: "bottom-0 right-0 size-2.5 ring-2",
	},
	lg: {
		px: 40,
		avatarClass: "size-10",
		textClass: "text-sm",
		indicatorClass: "bottom-0 right-0 size-3 ring-2",
	},
	xl: {
		px: 72,
		avatarClass: "size-[72px]",
		textClass: "text-lg",
		indicatorClass: "bottom-0.5 right-0.5 size-4 ring-[3px]",
	},
	"2xl": {
		px: 88,
		avatarClass: "size-[88px]",
		textClass: "text-2xl",
		indicatorClass: "bottom-1 right-1 size-[18px] ring-[3.5px]",
	},
};

export const UserProfileIndicator = memo(function UserProfileIndicator({
	status,
	className,
}: {
	status: UserPresenceStatus | "offline";
	className?: string;
}) {
	const resolvedStatus: UserPresenceStatus = status === "offline" ? "invisible" : status;
	const meta = USER_STATUS_BY_ID[resolvedStatus];
	return (
		<span
			className={cn(
				"inline-flex items-center justify-center rounded-full",
				meta.dotClass,
				className,
			)}
			aria-hidden
		>
			{resolvedStatus === "dnd" ? <span className="h-[2px] w-1.5 rounded-full bg-card" /> : null}
			{resolvedStatus === "invisible" ? <span className="size-1.5 rounded-full bg-card" /> : null}
		</span>
	);
});

type UserAvatarProps = {
	src?: string | null;
	alt?: string;
	initials: string;
	size?: UserAvatarSize;
	status?: UserPresenceStatus | "offline";
	ringClassName?: string;
	statusRingClassName?: string;
	className?: string;
	imgClassName?: string;
};

const AVATAR_INITIALS_CLASS =
	"flex size-full items-center justify-center rounded-full bg-muted font-semibold text-muted-foreground";

const MAX_FAILED_AVATAR_URLS = 500;
const failedAvatarUrlCache = new Set<string>();

function markFailedAvatarUrl(url: string): void {
	if (!url) {
		return;
	}
	if (failedAvatarUrlCache.size >= MAX_FAILED_AVATAR_URLS) {
		const oldest = failedAvatarUrlCache.values().next().value;
		if (oldest !== undefined) {
			failedAvatarUrlCache.delete(oldest);
		}
	}
	failedAvatarUrlCache.add(url);
}

type UserAvatarMediaProps = {
	resolvedSrc: string;
	alt: string;
	styles: (typeof SIZE_STYLES)[UserAvatarSize];
	initials: string;
	ringClassName?: string;
	imgClassName?: string;
};

const UserAvatarMedia = memo(function UserAvatarMedia({
	resolvedSrc,
	alt,
	styles,
	initials,
	ringClassName,
	imgClassName,
}: UserAvatarMediaProps) {
	const [failedPrimarySrc, setFailedPrimarySrc] = useState<string | null>(() =>
		resolvedSrc && failedAvatarUrlCache.has(resolvedSrc) ? resolvedSrc : null,
	);
	const [failedFallbackSrc, setFailedFallbackSrc] = useState<string | null>(null);

	const fallbackSrc =
		resolvedSrc && failedPrimarySrc === resolvedSrc ? toSameOriginObjectPath(resolvedSrc) : null;
	const activeSrc =
		fallbackSrc && fallbackSrc !== resolvedSrc
			? failedFallbackSrc === fallbackSrc || failedAvatarUrlCache.has(fallbackSrc)
				? ""
				: fallbackSrc
			: failedPrimarySrc === resolvedSrc
				? ""
				: resolvedSrc;

	if (!activeSrc) {
		return (
			<span
				className={cn(AVATAR_INITIALS_CLASS, styles.textClass, ringClassName, imgClassName)}
				aria-hidden={alt === "" ? true : undefined}
			>
				{initials}
			</span>
		);
	}

	return (
		/* eslint-disable-next-line @next/next/no-img-element */
		<img
			src={activeSrc}
			alt={alt}
			width={styles.px}
			height={styles.px}
			decoding="async"
			referrerPolicy="no-referrer"
			draggable={false}
			onError={() => {
				markFailedAvatarUrl(activeSrc);
				if (activeSrc === resolvedSrc) {
					setFailedPrimarySrc(resolvedSrc);
				} else {
					setFailedFallbackSrc(activeSrc);
				}
			}}
			className={cn(
				"pointer-events-none size-full rounded-full bg-card object-cover select-none",
				ringClassName,
				imgClassName,
			)}
		/>
	);
});

/**
 * Shared circular user avatar: photo via object-store URL, or neutral initials fallback.
 */
export const UserAvatar = memo(function UserAvatar({
	src,
	alt = "",
	initials,
	size = "md",
	status,
	ringClassName,
	statusRingClassName = "ring-card",
	className,
	imgClassName,
}: UserAvatarProps) {
	const styles = SIZE_STYLES[size];
	const raw = (src ?? "").trim();
	const resolvedSrc = raw ? resolvePublicObjectUrl(raw) : "";

	return (
		<div className={cn("relative shrink-0 select-none", styles.avatarClass, className)}>
			<UserAvatarMedia
				key={resolvedSrc || "initials"}
				resolvedSrc={resolvedSrc}
				alt={alt}
				styles={styles}
				initials={initials}
				ringClassName={ringClassName}
				imgClassName={imgClassName}
			/>
			{status ? (
				<UserProfileIndicator
					status={status}
					className={cn("absolute z-10", styles.indicatorClass, statusRingClassName)}
				/>
			) : null}
		</div>
	);
});
