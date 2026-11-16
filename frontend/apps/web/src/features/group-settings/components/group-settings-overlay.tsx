"use client";

import { FullscreenSettingsOverlay } from "@/components/layout/fullscreen-settings/fullscreen-settings-overlay";
import { GroupSettingsShell } from "@/features/group-settings/components/group-settings-shell";
import {
	GroupSettingsUnsavedProvider,
	useGroupSettingsUnsaved,
} from "@/features/group-settings/lib/group-settings-unsaved-guard";
import type {
	GroupSettingsSection,
	GroupSettingsTarget,
} from "@/features/group-settings/lib/group-settings-target";

const GROUP_SETTINGS_TITLE_ID = "group-settings-dialog-title";

type GroupSettingsOverlayProps = {
	open: boolean;
	onClose: () => void;
	target: GroupSettingsTarget | null;
	section: GroupSettingsSection;
};

export function GroupSettingsOverlay({
	open,
	onClose,
	target,
	section,
}: GroupSettingsOverlayProps) {
	if (!open || !target) {
		return null;
	}

	return (
		<GroupSettingsUnsavedProvider>
			<GroupSettingsOverlayDialog target={target} section={section} onClose={onClose} />
		</GroupSettingsUnsavedProvider>
	);
}

function GroupSettingsOverlayDialog({
	target,
	section,
	onClose,
}: {
	target: NonNullable<GroupSettingsOverlayProps["target"]>;
	section: GroupSettingsSection;
	onClose: () => void;
}) {
	const { runIfAllowed } = useGroupSettingsUnsaved();

	return (
		<FullscreenSettingsOverlay
			open
			titleId={GROUP_SETTINGS_TITLE_ID}
			onRequestClose={() => runIfAllowed(onClose)}
		>
			<GroupSettingsShell target={target} section={section} onClose={() => runIfAllowed(onClose)} />
		</FullscreenSettingsOverlay>
	);
}
