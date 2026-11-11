"use client";

import { useEffect } from "react";
import { createPortal } from "react-dom";
import { GroupSettingsShell } from "@/features/group-settings/components/group-settings-shell";
import { groupSettingsOverlayClass } from "@/features/group-settings/lib/group-settings-chrome";
import {
	GroupSettingsUnsavedProvider,
	useGroupSettingsUnsaved,
} from "@/features/group-settings/lib/group-settings-unsaved-guard";
import type {
	GroupSettingsSection,
	GroupSettingsTarget,
} from "@/features/group-settings/lib/group-settings-target";
import { useMounted } from "@/hooks/use-mounted";

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
	const mounted = useMounted();

	useEffect(() => {
		if (!open) {
			return;
		}
		const prev = document.body.style.overflow;
		document.body.style.overflow = "hidden";
		return () => {
			document.body.style.overflow = prev;
		};
	}, [open]);

	if (!open || !target || !mounted) {
		return null;
	}

	return createPortal(
		<GroupSettingsUnsavedProvider>
			<GroupSettingsOverlayDialog target={target} section={section} onClose={onClose} />
		</GroupSettingsUnsavedProvider>,
		document.body,
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

	useEffect(() => {
		const onKeyDown = (event: KeyboardEvent) => {
			if (event.key === "Escape") {
				event.preventDefault();
				runIfAllowed(onClose);
			}
		};
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [onClose, runIfAllowed]);

	return (
		<div
			className={groupSettingsOverlayClass}
			role="dialog"
			aria-modal="true"
			aria-labelledby={GROUP_SETTINGS_TITLE_ID}
		>
			<GroupSettingsShell target={target} section={section} onClose={() => runIfAllowed(onClose)} />
		</div>
	);
}
