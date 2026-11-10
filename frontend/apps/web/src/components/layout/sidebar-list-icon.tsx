import Image from "next/image";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

/** List row icon column — glyph ~54% of tile (12px → 24px, 14px → 26px at lg). */
export const sidebarListIconSlotSizeClass = "h-6 w-6 shrink-0 lg:h-[26px] lg:w-[26px]";

/** Photo / initials DMs — circular frame (not used for text-channel # glyphs). */
export const sidebarListIconBoxClass = cn(
	"flex items-center justify-center overflow-hidden rounded-full border border-sidebar-border bg-sidebar-control shadow-none",
	sidebarListIconSlotSizeClass,
);

/** Text channels in the conversation header — circular tile on chat surface (token colors, not white puck). */
export const conversationHeaderIconBoxClass =
	"flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-full border border-sidebar-border bg-sidebar-control text-muted-foreground shadow-none";

export const channelTileDefaultToneClass =
	"border-[color-mix(in_oklab,var(--chat-tab-active)_24%,var(--sidebar-border))] bg-[color-mix(in_oklab,var(--chat-tab-active)_14%,var(--sidebar-control))]";

export const channelTileAnnouncementToneClass =
	"border-[color-mix(in_oklab,var(--presence-idle)_40%,var(--sidebar-border))] bg-[color-mix(in_oklab,var(--presence-idle)_22%,var(--sidebar-control))]";

export type ChannelTileTone = "default" | "announcement";

/** Sidebar list: circular tile + subtle tint (not quick-view icon-only). */
export function channelListGlyphSlotClass(tone: ChannelTileTone = "default") {
	return cn(
		sidebarListIconBoxClass,
		"text-muted-foreground",
		tone === "announcement" ? channelTileAnnouncementToneClass : channelTileDefaultToneClass,
		"group-data-[selected=true]/row:border-sidebar-border group-data-[selected=true]/row:bg-sidebar-accent group-data-[selected=true]/row:text-sidebar-foreground",
	);
}

/** Quick views / nav rows — glyph only, fixed width for alignment with channel tiles. */
export const sidebarNavGlyphSlotClass = cn(
	"flex items-center justify-center text-muted-foreground",
	sidebarListIconSlotSizeClass,
	"group-data-[selected=true]/row:text-sidebar-foreground",
);

/** Sidebar row glyphs (quick views + channel `#` inside the tile). */
export const sidebarNavGlyphIconClass = "h-3 w-3 lg:h-3.5 lg:w-3.5";

export const channelListGlyphIconClass = sidebarNavGlyphIconClass;

export type SidebarListIconVariant = "list" | "header";

function iconBoxClass(variant: SidebarListIconVariant) {
	return variant === "header" ? conversationHeaderIconBoxClass : sidebarListIconBoxClass;
}

export function SidebarListIconBox({
	children,
	className,
	variant = "list",
}: {
	children: ReactNode;
	className?: string;
	variant?: SidebarListIconVariant;
}) {
	return <span className={cn(iconBoxClass(variant), className)}>{children}</span>;
}

/** Channel glyph tile; quick views use `framed={false}` for icon-only. */
export function ChannelGlyphSlot({
	children,
	className,
	variant = "list",
	framed = true,
	tileTone = "default",
}: {
	children: ReactNode;
	className?: string;
	variant?: SidebarListIconVariant;
	/** `false` = no circular background (quick views). */
	framed?: boolean;
	tileTone?: ChannelTileTone;
}) {
	if (variant === "header") {
		return (
			<span className={cn(conversationHeaderIconBoxClass, channelTileDefaultToneClass, className)}>
				{children}
			</span>
		);
	}
	if (!framed) {
		return <span className={cn(sidebarNavGlyphSlotClass, className)}>{children}</span>;
	}
	return <span className={cn(channelListGlyphSlotClass(tileTone), className)}>{children}</span>;
}

export function SidebarListAvatar({
	src,
	alt = "",
	variant = "list",
}: {
	src: string;
	alt?: string;
	variant?: SidebarListIconVariant;
}) {
	const px = variant === "header" ? 32 : 26;
	return (
		<SidebarListIconBox className="p-0" variant={variant}>
			<Image src={src} alt={alt} width={px} height={px} className="h-full w-full object-cover" />
		</SidebarListIconBox>
	);
}

/** 1:1 DM when no photo is available (same 20×20 tile as channels). */
export function SidebarListInitialsAvatar({
	initials,
	variant = "list",
}: {
	initials: string;
	variant?: SidebarListIconVariant;
}) {
	const text = initials.replace(/\s/g, "").slice(0, 2).toUpperCase();
	return (
		<SidebarListIconBox
			variant={variant}
			className={cn(
				"border-sidebar-border bg-sidebar-accent p-0 font-semibold leading-none text-sidebar-accent-foreground",
				variant === "header" ? "text-[10px]" : "text-[8px]",
			)}
		>
			{text}
		</SidebarListIconBox>
	);
}
