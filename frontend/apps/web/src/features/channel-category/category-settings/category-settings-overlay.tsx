"use client";

import { FullscreenSettingsOverlay } from "@/components/layout/fullscreen-settings/fullscreen-settings-overlay";
import { CategorySettingsShell } from "@/features/channel-category/category-settings/category-settings-shell";
import type {
	CategorySettingsSection,
	CategorySettingsTarget,
} from "@/features/channel-category/category-settings/category-settings-sidebar";
import {
	GroupSettingsUnsavedProvider,
	useGroupSettingsUnsaved,
} from "@/features/group-settings/lib/group-settings-unsaved-guard";

const CATEGORY_SETTINGS_TITLE_ID = "category-settings-dialog-title";

type CategorySettingsOverlayProps = {
	open: boolean;
	onClose: () => void;
	target: CategorySettingsTarget | null;
	section: CategorySettingsSection;
};

export function CategorySettingsOverlay({
	open,
	onClose,
	target,
	section,
}: CategorySettingsOverlayProps) {
	if (!open || !target) {
		return null;
	}

	return (
		<GroupSettingsUnsavedProvider>
			<CategorySettingsOverlayDialog target={target} section={section} onClose={onClose} />
		</GroupSettingsUnsavedProvider>
	);
}

function CategorySettingsOverlayDialog({
	target,
	section,
	onClose,
}: {
	target: NonNullable<CategorySettingsOverlayProps["target"]>;
	section: CategorySettingsSection;
	onClose: () => void;
}) {
	const { runIfAllowed } = useGroupSettingsUnsaved();

	return (
		<FullscreenSettingsOverlay
			open
			titleId={CATEGORY_SETTINGS_TITLE_ID}
			onRequestClose={() => runIfAllowed(onClose)}
		>
			<CategorySettingsShell
				target={target}
				section={section}
				onClose={() => runIfAllowed(onClose)}
				onForceClose={onClose}
			/>
		</FullscreenSettingsOverlay>
	);
}
