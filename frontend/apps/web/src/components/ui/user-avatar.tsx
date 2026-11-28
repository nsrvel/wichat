"use client";

import { memo, useState, type CSSProperties } from "react";
import { resolvePublicObjectUrl, toSameOriginObjectPath } from "@/lib/public-object-url";
import { cn } from "@/lib/utils";
import { USER_STATUS_BY_ID, type UserPresenceStatus } from "@/lib/user-status";

export type UserAvatarSize = "xs" | "sm" | "md" | "lg" | "xl" | "2xl";

/** Presence dot diameter vs avatar (`px` from `SIZE_STYLES`). */
const PRESENCE_DOT_RATIO = 0.26;

/** Nudge right (fraction of dot); small negative = sliver hangs past the avatar box edge. */
const PRESENCE_CORNER_OFFSET_RIGHT = -0.06;

/** Hollow “invisible” center cutout diameter vs presence dot. */
const PRESENCE_INVISIBLE_INNER_RATIO = 0.52;

/** Idle crescent: bite vs dot; bite center up-left so the arc opens to the left. */
const PRESENCE_IDLE_BITE_RATIO = 0.92;
const PRESENCE_IDLE_BITE_OFFSET_X = 0.44;
const PRESENCE_IDLE_BITE_OFFSET_Y = 0.14;

/** Opaque cutout ring (Tailwind `ring-*` class → surface token). */
const PRESENCE_CUTOUT_RING_COLOR: Record<string, string> = {
	"ring-sidebar": "var(--sidebar)",
	"ring-card": "var(--card)",
	"ring-chat-surface": "var(--chat-surface)",
};

function presenceCutoutRingPx(dotPx: number): number {
	return Math.max(3, Math.round(dotPx * 0.3));
}

function presenceIndicatorLayout(avatarPx: number): { dotPx: number; style: CSSProperties } {
	const dotPx = avatarPx * PRESENCE_DOT_RATIO;
	return {
		dotPx,
		style: {
			width: dotPx,
			height: dotPx,
			right: dotPx * PRESENCE_CORNER_OFFSET_RIGHT,
			bottom: 0,
		},
	};
}

const SIZE_STYLES: Record<
	UserAvatarSize,
	{
		px: number;
		avatarClass: string;
		textClass: string;
	}
> = {
	xs: {
		px: 24,
		avatarClass: "size-6",
		textClass: "text-[10px]",
	},
	sm: {
		px: 32,
		avatarClass: "size-8",
		textClass: "text-xs",
	},
	md: {
		px: 36,
		avatarClass: "size-9",
		textClass: "text-xs",
	},
	lg: {
		px: 40,
		avatarClass: "size-10",
		textClass: "text-sm",
	},
	xl: {
		px: 72,
		avatarClass: "size-[72px]",
		textClass: "text-lg",
	},
	"2xl": {
		px: 88,
		avatarClass: "size-[88px]",
		textClass: "text-2xl",
	},
};

export const UserProfileIndicator = memo(function UserProfileIndicator({
	status,
	dotPx = 12,
	cutoutRingClassName,
	className,
	style,
}: {
	status: UserPresenceStatus | "offline";
	/** Dot diameter in px; avatar badges pass a value derived from `SIZE_STYLES.px`. */
	dotPx?: number;
	/** Opaque outer ring on avatar badges (`ring-sidebar`, `ring-card`, …). */
	cutoutRingClassName?: string;
	className?: string;
	style?: CSSProperties;
}) {
	const resolvedStatus: UserPresenceStatus = status === "offline" ? "invisible" : status;
	const meta = USER_STATUS_BY_ID[resolvedStatus];
	const cutoutColor = cutoutRingClassName
		? PRESENCE_CUTOUT_RING_COLOR[cutoutRingClassName]
		: undefined;
	const cutoutPx = cutoutColor ? presenceCutoutRingPx(dotPx) : 0;
	const sizeStyle: CSSProperties = {
		...(style ?? { width: dotPx, height: dotPx }),
		...(cutoutColor ? { boxShadow: `0 0 0 ${cutoutPx}px ${cutoutColor}` } : undefined),
	};
	const dndBarH = Math.max(2, dotPx * 0.22);
	const dndBarW = dotPx * 0.56;
	const invisiblePx = Math.max(2, Math.round(dotPx * PRESENCE_INVISIBLE_INNER_RATIO));
	const idleBitePx = Math.max(2, Math.round(dotPx * PRESENCE_IDLE_BITE_RATIO));
	const idleBiteOffsetX = -dotPx * PRESENCE_IDLE_BITE_OFFSET_X;
	const idleBiteOffsetY = -dotPx * PRESENCE_IDLE_BITE_OFFSET_Y;
	const needsStatusOverlay =
		resolvedStatus === "dnd" || resolvedStatus === "invisible" || resolvedStatus === "idle";
	return (
		<span
			className={cn(
				"inline-flex items-center justify-center rounded-full",
				needsStatusOverlay && "relative",
				resolvedStatus === "idle" && "overflow-hidden",
				meta.dotClass,
				className,
			)}
			style={sizeStyle}
			aria-hidden
		>
			{resolvedStatus === "idle" ? (
				<span
					className={cn("absolute rounded-full", !cutoutColor && "bg-background")}
					style={{
						width: idleBitePx,
						height: idleBitePx,
						left: "50%",
						top: "50%",
						transform: `translate(calc(-50% + ${idleBiteOffsetX}px), calc(-50% + ${idleBiteOffsetY}px))`,
						...(cutoutColor ? { backgroundColor: cutoutColor } : undefined),
					}}
				/>
			) : null}
			{resolvedStatus === "dnd" ? (
				<span
					className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full bg-background"
					style={{ height: dndBarH, width: dndBarW }}
				/>
			) : null}
			{resolvedStatus === "invisible" ? (
				<span
					className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full bg-background"
					style={{ width: invisiblePx, height: invisiblePx }}
				/>
			) : null}
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
	const presence = presenceIndicatorLayout(styles.px);
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
					dotPx={presence.dotPx}
					cutoutRingClassName={statusRingClassName}
					style={presence.style}
					className="absolute z-10"
				/>
			) : null}
		</div>
	);
});
