"use client";

import { useQuery } from "@tanstack/react-query";
import { FileText, Shield } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
	FullscreenSettingsSidebar,
	type FullscreenSettingsNavGroup,
} from "@/components/layout/fullscreen-settings/fullscreen-settings-sidebar";
import { DeleteChannelConfirmDialog } from "@/features/channel/channel-settings/delete-channel-confirm-dialog";
import { translateError } from "@/integrations/i18n/translate-error";
import { useDeleteGroupChannelMutation } from "@/services/channel/channel.mutation";
import * as channelQuery from "@/services/channel/channel.query";

export type ChannelSettingsSection = "overview" | "permissions";

export type ChannelSettingsTarget = {
	groupId: string;
	channelId: string;
	channelName: string;
	channelType: string;
};

type ChannelSettingsSidebarProps = {
	target: ChannelSettingsTarget;
	activeSection: ChannelSettingsSection;
	onSectionChange: (section: ChannelSettingsSection) => void;
	onDeleted: () => void;
};

export function ChannelSettingsSidebar({
	target,
	activeSection,
	onSectionChange,
	onDeleted,
}: ChannelSettingsSidebarProps) {
	const { t } = useTranslation();
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
	const deleteMutation = useDeleteGroupChannelMutation(target.groupId);
	const channelsQuery = useQuery(channelQuery.groupChannelsQueryOpts(target.groupId));
	const liveChannel = channelsQuery.data?.channels.find((c) => c.id === target.channelId);
	const channelName = liveChannel?.name ?? target.channelName;

	const handleConfirmDelete = () => {
		deleteMutation.mutate(target.channelId, {
			onSuccess: () => {
				setDeleteDialogOpen(false);
				toast.success(t("channel.channel_settings.deleted_toast", { name: channelName }));
				onDeleted();
			},
			onError: (err) => {
				toast.error(translateError(t, err));
			},
		});
	};

	const navGroups: FullscreenSettingsNavGroup[] = [
		{
			id: "channel-settings",
			items: [
				{
					id: "overview",
					label: t("channel.channel_settings.overview"),
					icon: FileText,
					selected: activeSection === "overview",
					onSelect: () => onSectionChange("overview"),
				},
				{
					id: "permissions",
					label: t("channel.channel_settings.permissions"),
					icon: Shield,
					selected: activeSection === "permissions",
					onSelect: () => onSectionChange("permissions"),
				},
			],
		},
	];

	return (
		<>
			<DeleteChannelConfirmDialog
				channelName={channelName}
				open={deleteDialogOpen}
				onOpenChange={setDeleteDialogOpen}
				onConfirm={handleConfirmDelete}
				confirmDisabled={deleteMutation.isPending}
			/>
			<FullscreenSettingsSidebar
				headerTitle={channelName}
				navLabel={t("channel.channel_settings.header_title")}
				groups={navGroups}
				destructiveAction={{
					label: t("channel.channel_settings.delete_channel"),
					onClick: () => setDeleteDialogOpen(true),
				}}
			/>
		</>
	);
}
