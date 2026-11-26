"use client";

import type { CSSProperties, ReactNode } from "react";
import { cn } from "@/lib/utils";
import {
	userProfileBannerHeightPx,
	USER_PROFILE_BODY_CLASS,
	USER_PROFILE_CARD_WIDTH_PX,
} from "@/lib/user-profile-card-layout";

type UserProfileCardShellProps = {
	widthPx?: number;
	bannerColor?: string;
	roundedTop?: boolean;
	className?: string;
	bodyClassName?: string;
	children: ReactNode;
};

export function UserProfileCardShell({
	widthPx = USER_PROFILE_CARD_WIDTH_PX,
	bannerColor,
	roundedTop = true,
	className,
	bodyClassName,
	children,
}: UserProfileCardShellProps) {
	const bannerStyle: CSSProperties = {
		height: userProfileBannerHeightPx(widthPx),
		...(bannerColor ? { backgroundColor: bannerColor } : undefined),
	};

	return (
		<div className={cn("overflow-visible", className)} style={{ width: widthPx }}>
			<div
				className={cn(
					"w-full shrink-0 overflow-hidden bg-popover",
					roundedTop && "rounded-t-2xl",
				)}
				style={bannerStyle}
			/>
			<div className={cn(USER_PROFILE_BODY_CLASS, bodyClassName)}>{children}</div>
		</div>
	);
}