"use client";

import type { ReactNode } from "react";
import { GroupSettingsCloseButton } from "@/features/group-settings/components/group-settings-close-button";
import {
	groupSettingsBodyOuterClass,
	groupSettingsCanvasClass,
	groupSettingsMainPaneClass,
	groupSettingsShellClass,
	groupSettingsSidebarPaneClass,
	groupSettingsSplitBackdropClass,
} from "@/features/group-settings/lib/group-settings-chrome";

type FullscreenSettingsShellProps = {
	titleId: string;
	srTitle: string;
	onClose: () => void;
	sidebar: ReactNode;
	children: ReactNode;
};

/** Shared two-pane split-backdrop shell for Group Settings and Category Settings. */
export function FullscreenSettingsShell({
	titleId,
	srTitle,
	onClose,
	sidebar,
	children,
}: FullscreenSettingsShellProps) {
	return (
		<div className={groupSettingsShellClass}>
			<h1 id={titleId} className="sr-only">
				{srTitle}
			</h1>
			<GroupSettingsCloseButton onClose={onClose} />
			<div className={groupSettingsSplitBackdropClass} aria-hidden />
			<div className={groupSettingsBodyOuterClass}>
				<div className={groupSettingsCanvasClass}>
					<div className={groupSettingsSidebarPaneClass}>{sidebar}</div>
					<div className={groupSettingsMainPaneClass} role="tabpanel">
						{children}
					</div>
				</div>
			</div>
		</div>
	);
}
