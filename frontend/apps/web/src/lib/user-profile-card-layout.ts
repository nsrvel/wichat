import type { UserAvatarSize } from "@/components/ui/user-avatar";

/** Canonical user profile card (dock popover) width — previews scale from this. */
export const USER_PROFILE_CARD_WIDTH_PX = 320;

/** Banner height at `USER_PROFILE_CARD_WIDTH_PX` (half overlaps 2xl avatar). */
export const USER_PROFILE_BANNER_HEIGHT_PX = 108;

export const USER_PROFILE_AVATAR_SIZE: UserAvatarSize = "2xl";

/** 2xl avatar is 88px — center sits on banner bottom edge. */
export const USER_PROFILE_AVATAR_OVERLAP_PX = 44;

export const USER_PROFILE_AVATAR_RING_CLASS = "ring-[6px] ring-card";

export const USER_PROFILE_BODY_CLASS = "px-5 pb-5";

export const USER_PROFILE_DISPLAY_NAME_CLASS =
	"truncate text-[17px] font-bold leading-tight tracking-tight text-foreground";

export const USER_PROFILE_META_CLASS = "mt-0.5 truncate text-[13px] font-medium text-muted-foreground";

export const USER_PROFILE_BIO_CLASS =
	"mt-2.5 whitespace-pre-wrap break-words text-xs leading-relaxed text-muted-foreground";

export function userProfileBannerHeightPx(cardWidthPx = USER_PROFILE_CARD_WIDTH_PX): number {
	return Math.round(USER_PROFILE_BANNER_HEIGHT_PX * (cardWidthPx / USER_PROFILE_CARD_WIDTH_PX));
}

export function userProfileAvatarOverlapPx(cardWidthPx = USER_PROFILE_CARD_WIDTH_PX): number {
	return Math.round(
		USER_PROFILE_AVATAR_OVERLAP_PX * (cardWidthPx / USER_PROFILE_CARD_WIDTH_PX),
	);
}