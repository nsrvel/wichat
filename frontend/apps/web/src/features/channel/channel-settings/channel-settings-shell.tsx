"use client";

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { FullscreenSettingsShell } from "@/components/layout/fullscreen-settings/fullscreen-settings-shell";
import { ChannelSettingsOverviewSection } from "@/features/channel/channel-settings/channel-settings-overview-section";
import { ChannelSettingsPermissionsSection } from "@/features/channel/channel-settings/channel-settings-permissions-section";
import {
	ChannelSettingsSidebar,
	type ChannelSettingsSection,
	type ChannelSettingsTarget,
} from "@/features/channel/channel-settings/channel-settings-sidebar";
import { useGroupSettingsUnsaved } from "@/features/group-settings/lib/group-settings-unsaved-guard";

const CHANNEL_SETTINGS_TITLE_ID = "channel-settings-dialog-title";

type ChannelSettingsShellProps = {
	target: ChannelSettingsTarget;
	section: ChannelSettingsSection;
	onClose: () => void;
	onForceClose: () => void;
};

export function ChannelSettingsShell({
	target,
	section,
	onClose,
	onForceClose,
}: ChannelSettingsShellProps) {
	const { t } = useTranslation();
	const { runIfAllowed } = useGroupSettingsUnsaved();

	const syncKey = `${target.groupId}:${target.channelId}:${section}`;
	const [sectionOverride, setSectionOverride] = useState<{
		key: string;
		section: ChannelSettingsSection;
	} | null>(null);
	const activeSection = sectionOverride?.key === syncKey ? sectionOverride.section : section;

	const handleSectionChange = (next: ChannelSettingsSection) => {
		if (next === activeSection) {
			return;
		}
		runIfAllowed(() => setSectionOverride({ key: syncKey, section: next }));
	};

	return (
		<FullscreenSettingsShell
			titleId={CHANNEL_SETTINGS_TITLE_ID}
			srTitle={`${t("channel.channel_settings.header_title")} — ${target.channelName}`}
			onClose={onClose}
			sidebar={
				<ChannelSettingsSidebar
					target={target}
					activeSection={activeSection}
					onSectionChange={handleSectionChange}
					onDeleted={onForceClose}
				/>
			}
		>
			{activeSection === "overview" ? (
				<ChannelSettingsOverviewSection target={target} />
			) : (
				<ChannelSettingsPermissionsSection target={target} />
			)}
		</FullscreenSettingsShell>
	);
}
