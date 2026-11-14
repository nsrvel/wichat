"use client";

import { useState } from "react";
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
		indicatorClass: "-bottom-0.5 -right-0.5 size-2.5 ring-[1.5px]",
	},
	sm: {
		px: 32,
		avatarClass: "size-8",
		textClass: "text-xs",
		indicatorClass: "-bottom-0.5 -right-0.5 size-2.5 ring-2",
	},
	md: {
		px: 36,
		avatarClass: "size-9",
		textClass: "text-xs",
		indicatorClass: "-bottom-0.5 -right-0.5 size-3.5 ring-2",
	},
	lg: {
		px: 40,
		avatarClass: "size-10",
		textClass: "text-sm",
		indicatorClass: "-bottom-0.5 -right-0.5 size-3.5 ring-2",
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
		indicatorClass: "bottom-0.5 right-0.5 size-5 ring-[4px]",
	},
};

export function UserProfileIndicator({
	status,
	className,
}: {
	status: UserPresenceStatus;
	className?: string;
}) {
	const meta = USER_STATUS_BY_ID[status];
	return (
		<span
			className={cn(
				"inline-flex items-center justify-center rounded-full",
				meta.dotClass,
				className,
			)}
			aria-hidden
		>
			{status === "dnd" ? <span className="h-[2px] w-1.5 rounded-full bg-card" /> : null}
			{status === "invisible" ? <span className="size-1.5 rounded-full bg-card" /> : null}
		</span>
	);
}

type UserAvatarProps = {
	src?: string | null;
	alt?: string;
	initials: string;
	size?: UserAvatarSize;
	status?: UserPresenceStatus;
	ringClassName?: string;
	statusRingClassName?: string;
	className?: string;
	imgClassName?: string;
};

/**
 * Shared circular user avatar matching the UserProfilePopover style:
 * rounded-full photo or primary-tinted initials fallback, with optional presence indicator.
 */
export function UserAvatar({
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
	const trimmedSrc = src?.trim() || "";
	const [failedSrc, setFailedSrc] = useState<string | null>(null);
	const showImage = Boolean(trimmedSrc) && failedSrc !== trimmedSrc;

	return (
		<div className={cn("relative shrink-0", styles.avatarClass, className)}>
			{showImage ? (
				/* eslint-disable-next-line @next/next/no-img-element */
				<img
					src={trimmedSrc}
					alt={alt}
					width={styles.px}
					height={styles.px}
					onError={() => setFailedSrc(trimmedSrc)}
					className={cn(
						"size-full rounded-full bg-card object-cover",
						ringClassName,
						imgClassName,
					)}
				/>
			) : (
				<span
					className={cn(
						"flex size-full items-center justify-center rounded-full bg-primary/15 font-semibold text-primary",
						styles.textClass,
						ringClassName,
						imgClassName,
					)}
					aria-hidden={alt === "" ? true : undefined}
				>
					{initials}
				</span>
			)}
			{status ? (
				<UserProfileIndicator
					status={status}
					className={cn("absolute", styles.indicatorClass, statusRingClassName)}
				/>
			) : null}
		</div>
	);
}
