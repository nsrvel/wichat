import { AtSign, Bell, Cloud, Compass, FileText, Hash, ListTodo, MessageCircle, Server, Tv } from "lucide-react";
import type { ReactNode } from "react";
import type { ShellNavTarget } from "@/components/layout/shell-navigation";
import {
	SidebarListAvatar,
	ChannelGlyphSlot,
	SidebarListInitialsAvatar,
	channelListGlyphIconClass,
	conversationHeaderIconSlotSizeClass,
	sidebarNavGlyphSlotClass,
	type SidebarListIconVariant,
} from "@/components/layout/sidebar-list-icon";
import { cn } from "@/lib/utils";
import { ChannelListIcon } from "@/features/channel/components/channel-list-icon";

export const channelSidebarGlyphClass = channelListGlyphIconClass;
const channelHeaderGlyphClass = "h-4 w-4";

function glyphClass(variant: SidebarListIconVariant) {
	return variant === "header" ? channelHeaderGlyphClass : channelSidebarGlyphClass;
}

function wicloudGlyph(variant: SidebarListIconVariant) {
	const size = variant === "header" ? "h-4 w-4" : "h-[11px] w-[11px]";
	return <Cloud className={`${size} fill-[#F38020] text-[#F38020]`} strokeWidth={2} />;
}

function channelGlyph(id: string, variant: SidebarListIconVariant): ReactNode {
	const g = glyphClass(variant);
	switch (id) {
		case "ch-infra":
			return <Server className={g} strokeWidth={2.2} />;
		case "ch-wichat":
			return <MessageCircle className={g} strokeWidth={2.2} />;
		case "ch-wicloud":
			return wicloudGlyph(variant);
		case "ch-witask":
			return <ListTodo className={g} strokeWidth={2.2} />;
		case "ch-winobar":
			return <Tv className={g} strokeWidth={2.2} />;
		case "fav-announcements":
		case "fav-deployment-production":
		case "fav-deployment-staging":
			return <Bell className={g} strokeWidth={2.2} />;
		case "fav-release-notes":
			return <Hash className={g} strokeWidth={2.2} />;
		default:
			return <Hash className={g} strokeWidth={2.2} />;
	}
}

export const channelNavHashIcon = channelGlyph("fav-release-notes", "list");
export const channelNavBellIcon = channelGlyph("fav-announcements", "list");
export const channelNavInfraIcon = channelGlyph("ch-infra", "list");
export const channelNavWichatIcon = channelGlyph("ch-wichat", "list");
export const channelNavWicloudIcon = channelGlyph("ch-wicloud", "list");
export const channelNavWitaskIcon = channelGlyph("ch-witask", "list");
export const channelNavWinobarIcon = channelGlyph("ch-winobar", "list");

const DM_AVATAR_BY_ID: Record<string, string> = {
	"fav-self-dm": "/avatars/demo/marvell.png",
	"chat-dm-koh-yadi": "/avatars/demo/koh-yadi.jpg",
	"chat-dm-koh-yanto": "/avatars/demo/koh-yanto.jpg",
};

type ConversationNavIconProps = {
	active: ShellNavTarget;
	/** `header` = 32px tile; `list` = sidebar row size. */
	variant?: SidebarListIconVariant;
};

/** Sidebar list tile — larger only in conversation header. */
function quickViewGlyph(id: string, variant: SidebarListIconVariant): ReactNode {
	const g = glyphClass(variant);
	switch (id) {
		case "qv-mentions":
			return <AtSign className={g} strokeWidth={2} />;
		case "qv-discover":
			return <Compass className={g} strokeWidth={2} />;
		case "qv-drafts":
			return <FileText className={g} strokeWidth={2} />;
		default:
			return <AtSign className={g} strokeWidth={2} />;
	}
}

export function ConversationNavIcon({ active, variant = "list" }: ConversationNavIconProps) {
	if (active.kind === "quickView") {
		if (variant === "header") {
			return (
				<span
					className={cn(
						sidebarNavGlyphSlotClass,
						conversationHeaderIconSlotSizeClass,
						"text-muted-foreground",
					)}
				>
					{quickViewGlyph(active.id, variant)}
				</span>
			);
		}
		return (
			<ChannelGlyphSlot variant={variant} framed={false}>
				{quickViewGlyph(active.id, variant)}
			</ChannelGlyphSlot>
		);
	}

	if (active.kind === "dm") {
		const src = DM_AVATAR_BY_ID[active.id];
		if (src) {
			return <SidebarListAvatar src={src} alt={active.label} variant={variant} />;
		}
		return <SidebarListInitialsAvatar initials={active.label} variant={variant} />;
	}

	if (active.kind === "channel") {
		return (
			<ChannelListIcon
				type={active.channelType ?? ""}
				channelName={active.channelName ?? active.label}
				imageUrl={active.imageUrl}
				variant={variant}
			/>
		);
	}

	return <ChannelGlyphSlot variant={variant}>{channelGlyph(active.id, variant)}</ChannelGlyphSlot>;
}
