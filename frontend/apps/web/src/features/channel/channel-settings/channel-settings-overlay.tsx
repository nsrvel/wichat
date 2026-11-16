"use client";

import { FullscreenSettingsOverlay } from "@/components/layout/fullscreen-settings/fullscreen-settings-overlay";
import { ChannelSettingsShell } from "@/features/channel/channel-settings/channel-settings-shell";
import type {
	ChannelSettingsSection,
	ChannelSettingsTarget,
} from "@/features/channel/channel-settings/channel-settings-sidebar";
import {
	GroupSettingsUnsavedProvider,
	useGroupSettingsUnsaved,
} from "@/features/group-settings/lib/group-settings-unsaved-guard";

const CHANNEL_SETTINGS_TITLE_ID = "channel-settings-dialog-title";

type ChannelSettingsOverlayProps = {
	open: boolean;
	onClose: () => void;
	target: ChannelSettingsTarget | null;
	section: ChannelSettingsSection;
};

export function ChannelSettingsOverlay({
	open,
	onClose,
	target,
	section,
}: ChannelSettingsOverlayProps) {
	if (!open || !target) {
		return null;
	}

	return (
		<GroupSettingsUnsavedProvider>
			<ChannelSettingsOverlayDialog target={target} section={section} onClose={onClose} />
		</GroupSettingsUnsavedProvider>
	);
}

function ChannelSettingsOverlayDialog({
	target,
	section,
	onClose,
}: {
	target: NonNullable<ChannelSettingsOverlayProps["target"]>;
	section: ChannelSettingsSection;
	onClose: () => void;
}) {
	const { runIfAllowed } = useGroupSettingsUnsaved();

	return (
		<FullscreenSettingsOverlay
			open
			titleId={CHANNEL_SETTINGS_TITLE_ID}
			onRequestClose={() => runIfAllowed(onClose)}
		>
			<ChannelSettingsShell
				target={target}
				section={section}
				onClose={() => runIfAllowed(onClose)}
				onForceClose={onClose}
			/>
		</FullscreenSettingsOverlay>
	);
}
