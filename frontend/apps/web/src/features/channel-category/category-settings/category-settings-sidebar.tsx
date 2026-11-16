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
import { DeleteCategoryConfirmDialog } from "@/features/channel-category/category-settings/delete-category-confirm-dialog";
import { translateError } from "@/integrations/i18n/translate-error";
import { useDeleteGroupCategoryMutation } from "@/services/channel-category/channel-category.mutation";
import * as channelQuery from "@/services/channel/channel.query";

export type CategorySettingsSection = "overview" | "permissions";

export type CategorySettingsTarget = {
	groupId: string;
	categoryId: string;
	categoryName: string;
	isPrivate?: boolean;
};

type CategorySettingsSidebarProps = {
	target: CategorySettingsTarget;
	activeSection: CategorySettingsSection;
	onSectionChange: (section: CategorySettingsSection) => void;
	onDeleted: () => void;
};

export function CategorySettingsSidebar({
	target,
	activeSection,
	onSectionChange,
	onDeleted,
}: CategorySettingsSidebarProps) {
	const { t } = useTranslation();
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
	const deleteMutation = useDeleteGroupCategoryMutation(target.groupId);
	const channelsQuery = useQuery(channelQuery.groupChannelsQueryOpts(target.groupId));
	const liveCategory = channelsQuery.data?.categories.find((c) => c.id === target.categoryId);
	const categoryName = liveCategory?.name ?? target.categoryName;

	const handleConfirmDelete = () => {
		deleteMutation.mutate(target.categoryId, {
			onSuccess: () => {
				setDeleteDialogOpen(false);
				toast.success(t("channel_category.settings.deleted_toast", { name: categoryName }));
				onDeleted();
			},
			onError: (err) => {
				toast.error(translateError(t, err));
			},
		});
	};

	const navGroups: FullscreenSettingsNavGroup[] = [
		{
			id: "category-settings",
			items: [
				{
					id: "overview",
					label: t("channel_category.settings.overview"),
					icon: FileText,
					selected: activeSection === "overview",
					onSelect: () => onSectionChange("overview"),
				},
				{
					id: "permissions",
					label: t("channel_category.settings.permissions"),
					icon: Shield,
					selected: activeSection === "permissions",
					onSelect: () => onSectionChange("permissions"),
				},
			],
		},
	];

	return (
		<>
			<DeleteCategoryConfirmDialog
				categoryName={categoryName}
				open={deleteDialogOpen}
				onOpenChange={setDeleteDialogOpen}
				onConfirm={handleConfirmDelete}
				confirmDisabled={deleteMutation.isPending}
			/>
			<FullscreenSettingsSidebar
				headerTitle={categoryName}
				navLabel={t("channel_category.settings.header_title")}
				groups={navGroups}
				destructiveAction={{
					label: t("channel_category.settings.delete_category"),
					onClick: () => setDeleteDialogOpen(true),
				}}
			/>
		</>
	);
}
