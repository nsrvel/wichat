import { Hash } from "lucide-react";
import {
	ChannelGlyphSlot,
	SidebarListAvatar,
	channelHeaderGlyphIconClass,
	channelListGlyphIconClass,
	type ChannelTileTone,
	type SidebarListIconVariant,
} from "@/components/layout/sidebar-list-icon";
import {
	channelTileToneForType,
	resolveChannelIconSrc,
} from "@/features/channel/lib/channel-preset-icon";

type ChannelListIconProps = {
	type: string;
	channelName?: string;
	imageUrl?: string;
	variant?: SidebarListIconVariant;
	tileTone?: ChannelTileTone;
};

export function ChannelListIcon({
	type,
	imageUrl = "",
	variant = "list",
	tileTone,
}: ChannelListIconProps) {
	const resolvedTone = tileTone ?? channelTileToneForType(type);
	const presetSrc = resolveChannelIconSrc({ type, imageUrl });
	if (presetSrc) {
		return <SidebarListAvatar src={presetSrc} alt="" variant={variant} />;
	}

	const glyphClass = variant === "header" ? channelHeaderGlyphIconClass : channelListGlyphIconClass;

	return (
		<ChannelGlyphSlot variant={variant} tileTone={resolvedTone}>
			<Hash className={glyphClass} strokeWidth={2} />
		</ChannelGlyphSlot>
	);
}
